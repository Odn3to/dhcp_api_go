package controllers

import (
	
	"github.com/gofiber/fiber/v2"

	"os/exec"
	"fmt"
    "encoding/json"
    "strings"
    "encoding/csv"
    "os"
    "net/http"
    "bytes"

    "dhcp-api-go/database"
    "dhcp-api-go/resources/dhcp"
)

// @Summary Status do Serviço Kea DHCP
// @Description Retorna o Status do serviço Kea DHCP
// @ID status
// @Success 200 {object} dhcp.RetornoStatus
// @Router /dhcp/status [get]
func Status(c *fiber.Ctx) error {
    cmd := exec.Command("systemctl", "is-active", "kea-dhcp4-server.service")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Erro ao pegar status do kea", err)
        return err
    }

    text := "NÃO ATIVADO"
    classText := "alert-red"

    if string(output) == "active\n" {
        text = "ATIVADO"
        classText = "alert-green"
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "text":  text,
        "class": classText,
    })
}

// @Summary Get config DHCP
// @Description Retorna as configurações do DHCP
// @ID getConfigDHCP
// @Success 200 {object} dhcp.DataNetworkConfigurations
// @Router /dhcp/initConf [get]
func GetConfigDHCP(c *fiber.Ctx) error {
    var networkConfigurations dhcp.NetworkConfigurations
    var dataNetworkConfigurations dhcp.DataNetworkConfigurations
    database.DB.Where("id = 1 ").Find(&networkConfigurations)

    if err := json.Unmarshal([]byte(networkConfigurations.Data), &dataNetworkConfigurations); err != nil {
        fmt.Println("Erro ao analisar JSON:", err)
        return err
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "data":  dataNetworkConfigurations,
    })
}

// @Summary Configurar DHCP
// @Description Realiza a configuração do DHCP e restarta o serviço Kea
// @ID configDhcp
// @Accept  json
// @Produce  json
// @Param   ConfigDhcp     body    dhcp.DataNetworkConfigurations     true        "Configurações do DHCP"
// @Success 200 {object} dhcp.RetornoConf
// @Router /dhcp/conf [post]
func ConfiguradorDHCP(c *fiber.Ctx) error {
    var data dhcp.DataNetworkConfigurations
    if err := c.BodyParser(&data); err != nil {
        return c.JSON(fiber.Error{
            Message: err.Error(),
            Code:    fiber.StatusBadRequest,
        })
    }

    // Converte a estrutura DataNetworkConfigurations para JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return c.JSON(fiber.Error{
            Message: err.Error(),
            Code:    fiber.StatusInternalServerError,
        })
    }
    
    dataKea := map[string]string{
        "netInterface": data.NetInterface,
		"gateway":      data.Gateway,
		"subNet":       data.SubNet,
		"rangeInicial": data.RangeInicial,
		"rangeFinal":   data.RangeFinal,
		"lease":        data.Lease,
		"primario":     data.DNSPrimario,
		"secundario":   data.DNSSecundario,
	}
    
	config := dhcp.ConfigureKea(dataKea)
    fileName := "/etc/kea/kea-dhcp4.conf"

    // Abrir o arquivo para escrita (cria se não existir)
    file, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Erro ao criar o arquivo:", err)
        return err
    }
    defer file.Close()

    // Serializar o JSON e escrevê-lo no arquivo
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(config); err != nil {
        fmt.Println("Erro ao escrever o JSON no arquivo:", err)
        return err
    }

    //restart service Kea
    cmd := exec.Command("systemctl", "restart", "kea-dhcp4-server.service")
    err = cmd.Run()
    if err != nil {
        fmt.Println("Erro ao restart Kea!", err)
        return err
    }

    errGateway := dhcp.ConfigureGateway(data.Gateway)
    if errGateway  != nil {
        fmt.Println("Erro ao configurar Gateway!", errGateway)
        return errGateway 
    }

    //Aplica a conf do Gateway
    cmd = exec.Command("netplan", "apply")
    err = cmd.Run()
    if err != nil {
        fmt.Println("Erro ao aplicar as configurações do netPlan!", err)
        return err
    }

    // Atualiza o registro no banco de dados
    db := database.DB
    if err := db.Model(&dhcp.NetworkConfigurations{}).Where("id = 1").Update("Data", string(jsonData)).Error; err != nil {
        return c.JSON(fiber.Error{
            Message: err.Error(),
            Code:    fiber.StatusInternalServerError,
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "data": data,
    })
}

// @Summary Get interfaces
// @Description Retorna as interfaces
// @ID getInterfaces
// @Success 200 {object} dhcp.RetornoConf
// @Router /dhcp/interfaces [get]
func GetInterfaces(c *fiber.Ctx) error {
    cmd := exec.Command("ls", "/sys/class/net")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Erro ao buscar as interfaces!", err)
        return err
    }

    // Converte a saída em uma lista de nomes de interface
    interfaces := strings.Fields(string(output))

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "data":  interfaces,
    })
}

// @Summary Get Leases
// @Description Retorna os Leases
// @ID getLeases
// @Success 200 {object} dhcp.RetornoConf
// @Router /dhcp/data [get]
func GetCsv(c *fiber.Ctx) error {
    path := "/var/lib/kea/kea-leases4.csv"
    file, err := os.Open(path)
    if err != nil {
        fmt.Println("Erro ao abrir o arquivo CSV:", err)
        return err
    }
    defer file.Close()
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        fmt.Println("Erro ao ler o arquivo CSV:", err)
        return err
    }
    
    var jsonData []map[string]interface{}
    // A primeira linha do CSV contém os nomes dos campos
    header := records[0]
    for _, record := range records[1:] {
        data := make(map[string]interface{})
        for i, field := range header {
            data[field] = record[i]
        }
        jsonData = append(jsonData, data)
    }

    // Retornar como JSON
    return c.Status(fiber.StatusOK).JSON(jsonData)
}

func TokenValidationMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")

    tokenData := dhcp.TokenBody{Token: token}
    jsonData, err := json.Marshal(tokenData)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error creating JSON body",
        })
    }
    
    body := bytes.NewReader(jsonData)

    resp, err := http.Post("http://172.23.58.10/auth/login/validador", "application/json", body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error verifying token",
		})
	}
	defer resp.Body.Close()

	// Se o token não for válido, retorne um erro
	if resp.StatusCode != fiber.StatusOK {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	return c.Next()
}









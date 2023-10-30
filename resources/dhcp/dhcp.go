package dhcp

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"os"
	"os/exec"
)

func ConfigureKea(data map[string]string) *KeaConfig {
	config := &KeaConfig{}
	config.Dhcp4.InterfacesConfig.Interfaces = []string{data["netInterface"]}
	config.Dhcp4.LeaseDatabase.Type = "memfile"
	config.Dhcp4.LeaseDatabase.Persist = true
	config.Dhcp4.LeaseDatabase.Name = "/var/lib/kea/kea-leases4.csv"
	config.Dhcp4.LeaseDatabase.LfcInterval = 3600
	config.Dhcp4.RenewTimer = 15840
	config.Dhcp4.RebindTimer = 27720
	
	leaseValue, err := strconv.Atoi(data["lease"])
	if err != nil {
		fmt.Printf("Erro ao converter 'lease' para inteiro: %v\n", err)
		// Trate o erro de conversão aqui, se necessário
	}

	config.Dhcp4.ValidLifetime = leaseValue
	config.Dhcp4.OptionData = []struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}{
		{
			Name: "domain-name-servers",
			Data: fmt.Sprintf("%s, %s", data["primario"], data["secundario"]),
		},
		{
			Name: "domain-search",
			Data: "templab.lan",
		},
	}
	config.Dhcp4.Subnet4 = []struct {
		Subnet     string `json:"subnet"`
		Pools      []struct {
			Pool string `json:"pool"`
		} `json:"pools"`
		OptionData []struct {
			Name string `json:"name"`
			Data string `json:"data"`
		} `json:"option-data"`
	}{
		{
			Subnet: data["subNet"],
			Pools: []struct {
				Pool string `json:"pool"`
			}{
				{
					Pool: fmt.Sprintf("%s - %s", data["rangeInicial"], data["rangeFinal"]),
				},
			},
			OptionData: []struct {
				Name string `json:"name"`
				Data string `json:"data"`
			}{
				{
					Name: "routers",
					Data: data["gateway"],
				},
			},
		},
	}
	return config
}

func ConfigureGateway(ip string) error {
	filePath := "/etc/netplan/01-network-manager-all.yaml"

	// Lê o conteúdo do arquivo
	_, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Erro ao ler o arquivo: %v", err)
	}

	// Conteúdo a ser adicionado/modificado
	conteudo := fmt.Sprintf(`network:
   version: 2
   renderer: NetworkManager
   ethernets:
     ens160:
       addresses: [172.23.58.10/24]
       nameservers:
         addresses: [8.8.8.8, 8.8.4.4]
       routes:
           - to: 0.0.0.0/0
             via: 172.23.0.1
     ens192:
       addresses: [%s/24]
       `, ip)

	// Escreve o conteúdo modificado de volta no arquivo
	if err := ioutil.WriteFile(filePath, []byte(conteudo), os.ModePerm); err != nil {
		return fmt.Errorf("Erro ao escrever no arquivo: %v", err)
	}

	// Executa o comando 'sudo netplan apply'
	command := exec.Command("sudo", "netplan", "apply")
	if err := command.Run(); err != nil {
		return fmt.Errorf("Erro ao executar o comando: %v", err)
	}

	return nil
}
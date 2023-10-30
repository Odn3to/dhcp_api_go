package dhcp

// swagger: RetornoStatus
type RetornoStatus struct {
    Text string 
    Class string
}

// swagger: RetornoConf
type RetornoConf struct {
    Data string 
}

type TokenBody struct {
	Token string `json:"token"`
}

type NetworkConfigurations struct {
    ID uint     `json:"id"`
    Data string `json:"data"`
}

type DataNetworkConfigurations struct {
    NetInterface    string `json:"netInterface"`
    Gateway         string `json:"gateway"`
    SubNet          string `json:"subNet"`
    RangeInicial    string `json:"rangeInicial"`
    RangeFinal      string `json:"rangeFinal"`
    Lease           string `json:"lease"`
    DNSPrimario     string `json:"dnsPrimario"`
    DNSSecundario   string `json:"dnsSecundario"`
}

type KeaConfig struct {
	Dhcp4 struct {
		InterfacesConfig struct {
			Interfaces []string `json:"interfaces"`
		} `json:"interfaces-config"`
		LeaseDatabase struct {
			Type        string `json:"type"`
			Persist     bool   `json:"persist"`
			Name        string `json:"name"`
			LfcInterval int    `json:"lfc-interval"`
		} `json:"lease-database"`
		RenewTimer  int `json:"renew-timer"`
		RebindTimer int `json:"rebind-timer"`
		ValidLifetime int `json:"valid-lifetime"`
		OptionData  []struct {
			Name string `json:"name"`
			Data string `json:"data"`
		} `json:"option-data"`
		Subnet4 []struct {
			Subnet     string `json:"subnet"`
			Pools      []struct {
				Pool string `json:"pool"`
			} `json:"pools"`
			OptionData []struct {
				Name string `json:"name"`
				Data string `json:"data"`
			} `json:"option-data"`
		} `json:"subnet4"`
	} `json:"Dhcp4"`
}


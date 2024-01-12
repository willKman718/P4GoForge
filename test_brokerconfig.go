package main

import (
	"fmt"
	"os"
	"text/template"
)

type BrokerConfig struct {
	TargetPort  string
	ListenPort  string
	Directory   string
	Logfile     string
	AdminName   string
	AdminPhone  string
	AdminEmail  string
	ServerID    string
	ZaplockPath string
}

const brokerConfigTemplate = `# Sample p4broker configuration file

target      = {{.TargetPort}};
listen      = {{.ListenPort}};
directory   = {{.Directory}};
logfile     = "{{.Logfile}}";
debug-level = server=1;
admin-name  = "{{.AdminName}}";
admin-phone = {{.AdminPhone}};
admin-email = {{.AdminEmail}};
server.id   = {{.ServerID}};

command: ^zaplock$
{
    action = filter;
    checkauth = true;
    execute = {{.ZaplockPath}};
}`

func setupP4Broker(p4Test *P4Test) error {
	//rshp4dCommand := fmt.Sprintf(`"rsh:%s -r %s -L %s -d -q"`, p4Test.p4d, p4Test.serverRoot, p4Test.serverLog)
	p4Test.rshp4brokerCommand = fmt.Sprintf(`"rsh:%s -c %s -d -q"`, p4Test.p4broker, p4Test.brokerRoot+"/p4broker.cfg")

	// Broker configuration
	brokerConfig := BrokerConfig{
		TargetPort:  p4Test.rshp4dCommand,
		ListenPort:  p4Test.rshp4brokerCommand,
		Directory:   p4Test.binDir,
		Logfile:     p4Test.brokerRoot + "/p4broker.log",
		AdminName:   "Helix Core Admins",
		AdminPhone:  "999/911",
		AdminEmail:  "helix-core-admins@example.com",
		ServerID:    "brokerSvrID",
		ZaplockPath: p4Test.binDir + "/zaplock",
	}
	configPath := p4Test.brokerRoot + "/p4broker.cfg"

	if err := generateBrokerConfig(brokerConfig, configPath); err != nil {
		return fmt.Errorf("failed to generate broker config: %v", err)
	}
	if err := startP4broker(p4Test); err != nil {
		return fmt.Errorf("failed to start p4broker: %v", err)
	}

	return nil
}
func generateBrokerConfig(config BrokerConfig, filePath string) error {
	t := template.Must(template.New("brokerConfig").Parse(brokerConfigTemplate))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, config)
}

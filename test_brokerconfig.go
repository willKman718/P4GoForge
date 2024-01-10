package main

import (
	"html/template"
	"os"
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

func generateBrokerConfig(config BrokerConfig, filePath string) error {
	t := template.Must(template.New("brokerConfig").Parse(brokerConfigTemplate))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, config)
}

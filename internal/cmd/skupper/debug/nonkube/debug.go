package nonkube

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/internal/cmd/skupper/common"
	"github.com/skupperproject/skupper/internal/cmd/skupper/common/utils"
	"github.com/skupperproject/skupper/internal/config"
	internalclient "github.com/skupperproject/skupper/internal/nonkube/client/compat"
	"github.com/skupperproject/skupper/internal/nonkube/client/fs"
	"github.com/skupperproject/skupper/internal/utils/validator"
	"github.com/skupperproject/skupper/pkg/nonkube/api"
	"github.com/spf13/cobra"
)

type CmdDebug struct {
	CobraCmd            *cobra.Command
	Flags               *common.CommandDebugFlags
	namespace           string
	fileName            string
	connectorHandler    *fs.ConnectorHandler
	listenerHandler     *fs.ListenerHandler
	siteHandler         *fs.SiteHandler
	linkHandler         *fs.LinkHandler
	routerAccessHandler *fs.RouterAccessHandler
	certificateHandler  *fs.CertificateHandler
	secretHandler       *fs.SecretHandler
	configMapHandler    *fs.ConfigMapHandler
}

func NewCmdDebug() *CmdDebug {

	skupperCmd := CmdDebug{}

	return &skupperCmd
}

func (cmd *CmdDebug) NewClient(cobraCommand *cobra.Command, args []string) {
	if cmd.CobraCmd != nil && cmd.CobraCmd.Flag(common.FlagNameNamespace) != nil && cmd.CobraCmd.Flag(common.FlagNameNamespace).Value.String() != "" {
		cmd.namespace = cmd.CobraCmd.Flag(common.FlagNameNamespace).Value.String()
	}

	cmd.connectorHandler = fs.NewConnectorHandler(cmd.namespace)
	cmd.listenerHandler = fs.NewListenerHandler(cmd.namespace)
	cmd.siteHandler = fs.NewSiteHandler(cmd.namespace)
	cmd.linkHandler = fs.NewLinkHandler(cmd.namespace)
	cmd.routerAccessHandler = fs.NewRouterAccessHandler(cmd.namespace)
	cmd.certificateHandler = fs.NewCertificateHandler(cmd.namespace)
	cmd.secretHandler = fs.NewSecretHandler(cmd.namespace)
	cmd.configMapHandler = fs.NewConfigMapHandler(cmd.namespace)
}

func (cmd *CmdDebug) ValidateInput(args []string) error {
	var validationErrors []error
	fileStringValidator := validator.NewFilePathStringValidator()

	// Validate dump file name
	if len(args) < 1 {
		cmd.fileName = "skupper-dump"
	} else if len(args) > 1 {
		validationErrors = append(validationErrors, fmt.Errorf("only one argument is allowed for this command"))
	} else if args[0] == "" {
		validationErrors = append(validationErrors, fmt.Errorf("filename must not be empty"))
	} else {
		ok, err := fileStringValidator.Evaluate(args[0])
		if !ok {
			validationErrors = append(validationErrors, fmt.Errorf("filename is not valid: %s", err))
		} else {
			cmd.fileName = args[0]
		}
	}

	return errors.Join(validationErrors...)
}

func (cmd *CmdDebug) InputToOptions() {
	datetime := time.Now().Format("20060102150405")
	cmd.fileName = fmt.Sprintf("%s-%s-%s", cmd.fileName, cmd.namespace, datetime)
}

func (cmd *CmdDebug) Run() error {
	dumpFile := cmd.fileName

	// Add extension if not present
	if filepath.Ext(dumpFile) == "" {
		dumpFile = dumpFile + ".tar.gz"
	}

	tarFile, err := os.Create(dumpFile)
	if err != nil {
		return fmt.Errorf("Unable to save skupper dump details: %w", err)
	}

	// compress tar
	gz := gzip.NewWriter(tarFile)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	path := "/site-namespace/"
	rpath := path + "resources/"
	opts := fs.GetOptions{RuntimeFirst: true, LogWarning: false}
	platform := config.GetPlatform()

	if platform == types.PlatformPodman {
		pv, err := utils.RunCommand("podman", "version")
		if err == nil {
			utils.WriteTar("/versions/podman.txt", pv, time.Now(), tw)
		}
	}

	manifest, err := utils.RunCommand("skupper", "version", "-o", "yaml", "-n", cmd.namespace)
	if err == nil {
		utils.WriteTar("/versions/skupper.yaml", manifest, time.Now(), tw)
		utils.WriteTar("/versions/skupper.yaml.txt", manifest, time.Now(), tw)
	}

	//podman events ??
	//podman events --filter 84bebb4f09da

	sites, err := cmd.siteHandler.List(opts)
	if err == nil && sites != nil && len(sites) != 0 {
		for _, site := range sites {
			err := utils.WriteObject(site, rpath+"Site-"+site.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	routerAccesses, err := cmd.routerAccessHandler.List(opts)
	if err == nil && routerAccesses != nil && len(routerAccesses) != 0 {
		for _, routerAccess := range routerAccesses {
			err = utils.WriteObject(routerAccess, rpath+"RouterAccess-"+routerAccess.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	listeners, err := cmd.listenerHandler.List()
	if err == nil && listeners != nil && len(listeners) != 0 {
		for _, listener := range listeners {
			err := utils.WriteObject(listener, rpath+"Listener-"+listener.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	connectors, err := cmd.connectorHandler.List()
	if err == nil && connectors != nil && len(connectors) != 0 {
		for _, connector := range connectors {
			err := utils.WriteObject(connector, rpath+"Connector-"+connector.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	links, err := cmd.linkHandler.List(opts)
	if err == nil && links != nil && len(links) != 0 {
		for _, link := range links {
			err := utils.WriteObject(link, rpath+"Link-"+link.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	certificates, err := cmd.certificateHandler.List()
	if err == nil && certificates != nil && len(certificates) != 0 {
		for _, certificate := range certificates {
			err := utils.WriteObject(certificate, rpath+"Certificate-"+certificate.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	secrets, err := cmd.secretHandler.List()
	if err == nil && secrets != nil && len(secrets) != 0 {
		for _, secret := range secrets {
			err := utils.WriteObject(secret, rpath+"Secret-"+secret.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	configMaps, err := cmd.configMapHandler.List()
	if err == nil && configMaps != nil && len(configMaps) != 0 {
		for _, configMap := range configMaps {
			err := utils.WriteObject(configMap, rpath+"ConfigMap-"+configMap.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	skrPath := api.GetDefaultOutputPath(cmd.namespace) + "/" + string(api.RouterConfigPath+"/skrouterd.json")
	//skrPath := pathProvider.GetRuntimeNamespace()+ api.
	skrouterd, err := os.ReadFile(skrPath)
	if err == nil && skrouterd != nil {
		utils.WriteTar(rpath+"skrouterd.json", skrouterd, time.Now(), tw)
	}

	cli, err := internalclient.NewCompatClient(os.Getenv("CONTAINER_ENDPOINT"), "")
	if err == nil {
		rtrContainerName := cmd.namespace + "-skupper-router"
		if container, err := cli.ContainerInspect(rtrContainerName); err == nil {
			encodedOutput, _ := utils.Encode("yaml", container)
			utils.WriteTar(rpath+"Container-"+container.Name+".yaml", []byte(encodedOutput), time.Now(), tw)
		}

		out, err := cli.ContainerExec(rtrContainerName, strings.Split("skstat -c", " "))
		if err == nil {
			fmt.Println("containerexec: ", err, out)
		}

		logs, err := cli.ContainerLogs(rtrContainerName)
		if err == nil {
			utils.WriteTar(path+"logs/"+rtrContainerName+".txt", []byte(logs), time.Now(), tw)
		}

		ctlContainerName := "system-controller"
		if container, err := cli.ContainerInspect(ctlContainerName); err == nil {
			encodedOutput, _ := utils.Encode("yaml", container)
			utils.WriteTar(rpath+"Container-"+container.Name+".yaml", []byte(encodedOutput), time.Now(), tw)
		}

		out, err = cli.ContainerExec(ctlContainerName, strings.Split("skstat -c", " "))
		if err == nil {
			fmt.Println("containerexec: ", err, out)
		}

		logs, err = cli.ContainerLogs(ctlContainerName)
		if err == nil {
			utils.WriteTar(path+"logs/"+ctlContainerName+".txt", []byte(logs), time.Now(), tw)
		}
	}

	fmt.Println("Skupper dump details written to compressed archive: ", dumpFile)
	return nil
}

func (cmd *CmdDebug) WaitUntil() error { return nil }

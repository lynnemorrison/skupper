package nonkube

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	"github.com/skupperproject/skupper/pkg/generated/client/clientset/versioned/scheme"
	"github.com/skupperproject/skupper/pkg/nonkube/api"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
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
		pv, err := runCommand("podman", "version")
		if err == nil {
			writeTar("/versions/podman.txt", pv, time.Now(), tw)
		}
	}

	manifest, err := runCommand("skupper", "version", "-o", "yaml", "-n", cmd.namespace)
	if err == nil {
		writeTar("/versions/skupper.yaml", manifest, time.Now(), tw)
		writeTar("/versions/skupper.yaml.txt", manifest, time.Now(), tw)
	}

	//podman events ??
	//podman events --filter 84bebb4f09da

	sites, err := cmd.siteHandler.List(opts)
	if err == nil && sites != nil && len(sites) != 0 {
		for _, site := range sites {
			err := writeObject(site, rpath+"Site-"+site.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	routerAccesses, err := cmd.routerAccessHandler.List(opts)
	if err == nil && routerAccesses != nil && len(routerAccesses) != 0 {
		for _, routerAccess := range routerAccesses {
			err = writeObject(routerAccess, rpath+"RouterAccess-"+routerAccess.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	listeners, err := cmd.listenerHandler.List()
	if err == nil && listeners != nil && len(listeners) != 0 {
		for _, listener := range listeners {
			err := writeObject(listener, rpath+"Listener-"+listener.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	connectors, err := cmd.connectorHandler.List()
	if err == nil && connectors != nil && len(connectors) != 0 {
		for _, connector := range connectors {
			err := writeObject(connector, rpath+"Connector-"+connector.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	links, err := cmd.linkHandler.List(opts)
	if err == nil && links != nil && len(links) != 0 {
		for _, link := range links {
			err := writeObject(link, rpath+"Link-"+link.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	certificates, err := cmd.certificateHandler.List()
	if err == nil && certificates != nil && len(certificates) != 0 {
		for _, certificate := range certificates {
			err := writeObject(certificate, rpath+"Certificate-"+certificate.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	secrets, err := cmd.secretHandler.List()
	if err == nil && secrets != nil && len(secrets) != 0 {
		for _, secret := range secrets {
			err := writeObject(secret, rpath+"Secret-"+secret.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	configMaps, err := cmd.configMapHandler.List()
	if err == nil && configMaps != nil && len(configMaps) != 0 {
		for _, configMap := range configMaps {
			err := writeObject(configMap, rpath+"ConfigMap-"+configMap.Name, tw)
			if err != nil {
				return err
			}
		}
	}

	skrPath := api.GetDefaultOutputPath(cmd.namespace) + "/" + string(api.RouterConfigPath+"/skrouterd.json")
	//skrPath := pathProvider.GetRuntimeNamespace()+ api.
	skrouterd, err := os.ReadFile(skrPath)
	if err == nil && skrouterd != nil {
		writeTar(rpath+"skrouterd.json", skrouterd, time.Now(), tw)
	}

	cli, err := internalclient.NewCompatClient(os.Getenv("CONTAINER_ENDPOINT"), "")
	if err == nil {
		rtrContainerName := cmd.namespace + "-skupper-router"
		if container, err := cli.ContainerInspect(rtrContainerName); err == nil {
			encodedOutput, _ := utils.Encode("yaml", container)
			writeTar(rpath+"Container-"+container.Name+".yaml", []byte(encodedOutput), time.Now(), tw)
		}

		out, err := cli.ContainerExec(rtrContainerName, strings.Split("skstat -c", " "))
		if err == nil {
			fmt.Println("containerexec: ", err, out)
		}

		logs, err := cli.ContainerLogs(rtrContainerName)
		if err == nil {
			writeTar(path+"logs/"+rtrContainerName+".txt", []byte(logs), time.Now(), tw)
		}

		ctlContainerName := "system-controller"
		if container, err := cli.ContainerInspect(ctlContainerName); err == nil {
			encodedOutput, _ := utils.Encode("yaml", container)
			writeTar(rpath+"Container-"+container.Name+".yaml", []byte(encodedOutput), time.Now(), tw)
		}

		out, err = cli.ContainerExec(ctlContainerName, strings.Split("skstat -c", " "))
		if err == nil {
			fmt.Println("containerexec: ", err, out)
		}

		logs, err = cli.ContainerLogs(ctlContainerName)
		if err == nil {
			writeTar(path+"logs/"+ctlContainerName+".txt", []byte(logs), time.Now(), tw)
		}
	}

	fmt.Println("Skupper dump details written to compressed archive: ", dumpFile)
	return nil
}

func (cmd *CmdDebug) WaitUntil() error { return nil }

// helper functions
func runCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeTar(name string, data []byte, ts time.Time, tw *tar.Writer) error {
	hdr := &tar.Header{
		Name:    name,
		Mode:    0600,
		Size:    int64(len(data)),
		ModTime: ts,
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		return fmt.Errorf("Failed to write tar file header: %w", err)
	}
	_, err = tw.Write(data)
	if err != nil {
		return fmt.Errorf("Failed to write to tar archive: %w", err)
	}
	return nil
}

func writeObject(rto runtime.Object, name string, tw *tar.Writer) error {
	var b bytes.Buffer
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	if err := s.Encode(rto, &b); err != nil {
		return err
	}
	err := writeTar(name+".yaml", b.Bytes(), time.Now(), tw)
	if err != nil {
		return err
	}
	err = writeTar(name+".yaml.txt", b.Bytes(), time.Now(), tw)
	if err != nil {
		return err
	}
	return nil
}

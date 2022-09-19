package main

import (
    "errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    //cp "github.com/otiai10/copy"

    "github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/applysetters"

    "sigs.k8s.io/kustomize/kyaml/fn/framework"
    "sigs.k8s.io/kustomize/kyaml/fn/framework/command"
    "sigs.k8s.io/kustomize/kyaml/kio"
    "sigs.k8s.io/kustomize/kyaml/yaml"
)

var debugFp *os.File = nil

type UpfNadGen struct {
    yaml.ResourceMeta   `json:",inline" yaml:",inline"`
    UpfNadGenSpec  `json:"spec" yaml:"spec"`
}

type UpfNadGenSpec struct {
    N3Cni       string  `json:"n3cni,omitempty" yaml:"n3cni,omitempty"`
    N3Master    string  `json:"n3master,omitempty" yaml:"n3master,omitempty"`
    N3Gw        string  `json:"n3gw,omitempty" yaml:"n3gw,omitempty"`
    N4Cni       string  `json:"n4cni,omitempty" yaml:"n4cni,omitempty"`
    N4Master    string  `json:"n4master,omitempty" yaml:"n4master,omitempty"`
    N4Gw        string  `json:"n4gw,omitempty" yaml:"n4gw,omitempty"`
    N6Cni       string  `json:"n6cni,omitempty" yaml:"n6cni,omitempty"`
    N6Master    string  `json:"n6master,omitempty" yaml:"n6master,omitempty"`
    N6Gw        string  `json:"n6gw,omitempty" yaml:"n6gw,omitempty"`
}

var UpfNadCfg string = `{
     "cniVersion": "$CNI_VERSION"
      "plugins": [
        {
          "type": "$CNI_TYPE",
          "capabilities": { "ips": true },
          "master": "$NAD_MASTER",
          "mode": "bridge",
          "ipam": {
            "type": "static",
            "routes": [
              {
                "dst": "0.0.0.0/0",
                "gw": "$NAD_GW"
              }
            ]
          }
        }, {
          "capabilities": { "mac": true },
          "type": "tuning"
        }
    }`

func makeDirectory(path string) error {
    return os.MkdirAll(path, 0755)
}

func applySetterOnSingleFile(item *yaml.RNode, setters map[string]string) ([]*yaml.RNode, error) {
	var f applysetters.ApplySetters

	var s []applysetters.Setter
	for k, v := range setters {
		s = append(s, applysetters.Setter{k, v})
	}
	f.Setters = s

	ni, err := f.Filter([]*yaml.RNode{item})
	return ni, err
}

func applySetterAndWriteFile (path string, append *os.File, setters map[string]string, attachment string) {
	item,_ := yaml.ReadFile(path)
	items,_ := applySetterOnSingleFile(item, setters)
	str,_ := items[0].String()
	append.Write([]byte(str + "'" + attachment + "'"))
	append.Write([]byte("\n\n---\n\n"))
}

func (s *UpfNadGen) writeNadInfo(path string, srcPath string, nadName string, namespace string, interfaceType string) error {
    var setters map[string]string = make(map[string]string)
    var cni, master, gw string
    switch interfaceType {
    case "n3":
        cni = s.N3Cni
        master = s.N3Master
        gw = s.N3Gw
    case "n4":
        cni = s.N4Cni
        master = s.N4Master
        gw = s.N4Gw
    case "n6":
        cni = s.N6Cni
        master = s.N6Master
        gw = s.N6Gw
    default:
        return errors.New("Unsupported UPF interface type " + interfaceType)
    }
    nadcfg := strings.Clone(UpfNadCfg)
    setters["nadname"] = nadName + "-" + interfaceType
    setters["nadns"] = namespace
    nadcfg = strings.Replace(nadcfg, "$CNI_VERSION", "0.3.1", 1)
    nadcfg = strings.Replace(nadcfg, "$CNI_TYPE", cni, 1)
    nadcfg = strings.Replace(nadcfg, "$NAD_MASTER", master, 1)
    nadcfg = strings.Replace(nadcfg, "$NAD_GW", gw, 1)
    debug("nadcfg is " + nadcfg)
    fileName := path + "/" + interfaceType + ".yaml"
    debug("opening file " + fileName)
    fp, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Println("Error: open " + fileName + "failed :" + err.Error())
        return err
    }
    applySetterAndWriteFile(srcPath + "/templates/nad.yaml", fp, setters, nadcfg)
    return nil
}

func (gen *UpfNadGen) createDirs(basePath string, srcDir string, nadName string, namespace string) error {
    if err := makeDirectory(basePath); err != nil {
        log.Println("Failed to make directory " + basePath + " : " + err.Error())
        return err
    }

    debug("gen is " + fmt.Sprintf("%v", gen))
    debug(fmt.Sprintf("n3 CNI is %s", gen.N3Cni))
    if err := gen.writeNadInfo(basePath, srcDir, nadName, namespace, "n3"); err != nil {
        log.Println("Failed to write N3 Nad info: " + err.Error())
        return err
    }
    if err := gen.writeNadInfo(basePath, srcDir, nadName, namespace, "n4"); err != nil {
        log.Println("Failed to write N3 Nad info: " + err.Error())
        return err
    }
    if err := gen.writeNadInfo(basePath, srcDir, nadName, namespace, "n6"); err != nil {
        log.Println("Failed to write N3 Nad info: " + err.Error())
        return err
    }
    return nil
}

func processUpfNadGen(node *yaml.RNode, destDir string, srcDir string) error {
    upfnadgen := &UpfNadGen{}
    err := node.YNode().Decode(upfnadgen)
    if err != nil {
        return err
    }
    nadName := node.GetName()
    namespace := node.GetNamespace()
    basePath := filepath.Join(destDir, nadName)
    debug(fmt.Sprintf("nadName %s namespace %s basePath %s srcDir %s", nadName, namespace, basePath, srcDir))
    if err := upfnadgen.createDirs(basePath, srcDir, nadName, namespace); err != nil {
      log.Println("Failed to create and populate directories " + err.Error())
      return err
    }
    return nil
}

func filterRnodeByKind(kind string, items []*yaml.RNode) []*yaml.RNode {
    s := &framework.Selector{
      Kinds: []string{kind},
    }
    rets, _ := s.Filter(items)
    return rets
}

func debug(msg string) {
  debugFp.WriteString(msg + "\n")
  debugFp.Sync()
}

func main() {

    debugFp, _ = os.Create("/tmp/cnna-debug")

    defer debugFp.Close()

    var config struct {
        Data map[string]string `yaml:"data"`
    }

    fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
      rets := filterRnodeByKind("UpfNadGen", items)
      debug(fmt.Sprintf("UpfNadGen has %d Rnodes", len(rets)))
      if len(rets) == 0 {
        log.Println("Error: UpfNadGen resource not found")
        return items, nil
      }
      for _, item := range(rets) {
        err := processUpfNadGen(item, config.Data["destdir"], config.Data["srcdir"])
        if err != nil {
          log.Println("Failed to process UpfNadGen: " + err.Error())
          debug(fmt.Sprintf("Failed to process UpfNadGen: %s", err.Error()))
        }
      }
      return items, nil
    }

    p := framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: &config}
    cmd := command.Build(p, command.StandaloneDisabled, false)

    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}

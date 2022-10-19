package main

import (
    //"errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    //"strings"

    cp "github.com/otiai10/copy"

    //"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/applysetters"

    "sigs.k8s.io/kustomize/kyaml/fn/framework"
    "sigs.k8s.io/kustomize/kyaml/fn/framework/command"
    "sigs.k8s.io/kustomize/kyaml/kio"
    "sigs.k8s.io/kustomize/kyaml/yaml"
)

var debugFp *os.File = nil

type NfDeploy struct {
    yaml.ResourceMeta `json:",inline" yaml:",inline"`
    NfDeploySpec      `json:"spec" yaml:"spec"`
}

type NfDeploySpec struct {
    Sites   []NfDeloySite
}

type NfDeloySite struct {
    Id              string `json:"id,omitempty" yaml:"id,omitempty"`
    LocationName    string `json:"locationName,omitempty" yaml:"locationName,omitempty"`
    ClusterName     string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
    NfKind          string `json:"nfKind,omitempty" yaml:"nfKind,omitempty"`
    NfClassName     string `json:"nfClassName,omitempty" yaml:"nfClassName,omitempty"`
    NfVendor        string `json:"nfVendor,omitempty" yaml:"nfVendor,omitempty"`
    NfVersion       string `json:"nfVersion,omitempty" yaml:"nfVersion,omitempty"`
    NfNamespace     string `json:"nfNamespace,omitempty" yaml:"nfNamespace,omitempty"`
    Connectivities  []NfDeployConnectivity
}

type NfDeployConnectivity struct {
    NeighborName    string `json:"neighborName,omitempty" yaml:"neighborName,omitempty"`
    CapacityProfile string `json:"capacityProfile,omitempty" yaml:"capacityProfile,omitempty"`
}

func makeDirectory(path string) error {
    return os.MkdirAll(path, 0755)
}

func (s *NfDeloySite) populate(basePath string, deploymentName string, repoPath string) error {
    if err := makeDirectory(basePath); err != nil {
        log.Println("Failed to make directory " + basePath + " : " + err.Error())
        return err
    }

    // check vendor and type
    srcCommonPkg := repoPath + "/" + s.NfKind + "/common"
    dstCommonPkg := basePath + "/" + s.ClusterName + "/" + s.Id

    if err := cp.Copy(srcCommonPkg, dstCommonPkg); err != nil {
        return err
    }

    srcVendorPkg := repoPath + "/" + s.NfVendor + "/" + s.NfKind + "/" + s.NfVersion
    //dstVendorPkg := basePath + "/" + s.NfVendor + "/" s.NfKind + "/" + s.NfVersion

    if err := cp.Copy(srcVendorPkg, dstCommonPkg); err != nil {
        return err
    }

    return nil
}

func processNfDeploy(node *yaml.RNode, destDir string, repoPath string) error {
    nfDeploy := &NfDeploy{}
    err := node.YNode().Decode(nfDeploy)
    if err != nil {
        return err
    }
    dmName := node.GetName()
    //namespace := node.GetNamespace()
    basePath := filepath.Join(destDir, dmName)
    debug(fmt.Sprintf("dmName %s namespace %s basePath %s", dmName, basePath))
    for _, site := range nfDeploy.Sites {
        if err := site.populate(basePath, dmName, repoPath); err != nil {
            log.Println("Failed to create and populate directories " + err.Error())
            return err
        }
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
      rets := filterRnodeByKind("NfDeploy", items)
      debug(fmt.Sprintf("NfDeploy has %d Rnodes", len(rets)))
      if len(rets) == 0 {
        log.Println("Error: NfDeploy resource not found")
        return items, nil
      }
      for _, item := range(rets) {
        err := processNfDeploy(item, config.Data["destdir"], config.Data["repodir"])
        if err != nil {
          log.Println("Failed to process NfDeploy: " + err.Error())
          debug(fmt.Sprintf("Failed to process NfDeploy: %s", err.Error()))
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

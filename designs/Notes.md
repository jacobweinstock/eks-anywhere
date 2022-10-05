# Bootstrap Cluster Components

## Terms

### EKS-A Terms

|     |     |
| --- | --- |
|Bootstrap Component:| Requirements of CAPI needed to be able to submit |
|EKS-A Provider: | A client that manages all provider specific tasks needed when running a CAPI (via `clusterctl`) workflow. It is optionally responsible for deploying/managing an Infrastructure Provider. |

### CAPI Terms

|     |     |
| --- | --- |
|CAPI Provider:| A client/controller for managing [provider components][] and communicating with an Infrastructure Provider.|
|[Infrastructure Provider][]:| A source of computational resources, exposed via an API.|

## Components

* CAPI components
* EKS-A components
* EKS-A Provider components

## Current

Existing EKS-A providers: cloudstack, docker, snow, tinkerbell, vsphere

Currently, `func (s *CreateBootStrapClusterTask) Run` handles creating the bootstrap cluster and installing components. The following are just the functions that pertain to bootstrap cluster components.

* `commandContext.Provider.PreCAPIInstallOnBootstrap`
  * provider specific task
  * __Provider Usage__
    * [cloudstack] updates secrets here.
    * [tinkerbell] installs tinkerbell stack here.
* `commandContext.ClusterManager.InstallCAPI`
  * `c.clusterClient.InitInfrastructure`
    * `clusterctl.InitInfrastructure`
      * `clusterctl init`
  * `c.waitForCAPI`
    * `c.clusterClient.waitForDeployments`
      * `k.WaitForDeployment`
        * `k.Wait`
          * `k.wait`
            * `kubectl wait`
* `commandContext.ClusterManager.CreateAwsIamAuthCaSecret`
* `commandContext.Provider.PostBootstrapSetup`
  * provider specific task
  * __Provider Usage__
    * [tinkerbell] applies hardware spec and waits for BMC connectivity here.

## Questions

* Do we need a provider specific pre install CAPI task? Can all provider specific things be done after generic CAPI is installed?
* Why is CreateAwsIamAuthCaSecret needed in the bootstrap cluster? is this specific to a particular provider?
* Where(or do we?) do we write out the `.cluster-api/clusterctl.yaml` file?
* I need help understanding `retrierClient` and `clustermanager.client`. what cluster is `clustermanager` managing? kubernetes cluster? bootstrap cluster? CAPI and EKS-A enabled kubernetes cluster? and why is it an interface (`ClusterClient`)? are there multiple implementations of this interface? just for testing (not very test friendly being so large)?

## Proposed

* Create a dedicated function/method/task that just installs bootstrap cluster components, no installing of kubernetes with kind.
  1. run CAPI install
  2. run provider component install
  3. run EKS-A provider component install (potentially, but only if CreateAwsIamAuthCaSecret is really needed)
* dependencies/inputs
  * kubeconfig
  * clusterctl executable
  * kubectl executable
  * provider interface

---

[Infrastructure Provider]: https://cluster-api.sigs.k8s.io/reference/glossary.html#infrastructure-provider
[provider components]: https://cluster-api.sigs.k8s.io/reference/glossary.html#provider-components

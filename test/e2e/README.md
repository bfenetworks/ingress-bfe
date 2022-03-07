# BFE ingress controller e2e test

This test follows K8s project [ingress-controller-conformance](https://github.com/kubernetes-sigs/ingress-controller-conformance), and add more test cases for ingress-bfe special features.

## How to run

To run all e2e test cases, execute following command in ingress-bfe project's top directory:

``` 
$ make e2e-test
```
It would automatically start the whole testing with following procedures:

- Build bfe-ingress-controller docker image
- Prepare test environment, including spining up a local k8s cluster with [Kind](https://kind.sigs.k8s.io/), loading docker images, etc. All scripts used to prepare environment are located in [test/script](../script).
- Execute test cases by running [run.sh](./run.sh), which actually build and execute program e2e_test.

## Contributing

We encourage contributors write e2e test case for new feature needed to be merged into ingress-bfe.

The test code is based on BDD testing framework [godog](https://github.com/cucumber/godog), a testing framework of [cucumber](https://cucumber.io/). It uses [Gherkin Syntax]( https://cucumber.io/docs/gherkin/reference/) to describe test case.

Steps to add new test case as below:

### Step1: Create Gherkin feature

* Create feature file to describe your test case. 

All feature files are under directory [test/e2e/features](./features). Please put your feature file into proper sub-directory. For example, features/<your-new-feature>/<your-new-feature>.feature
  > Try to reuse steps from existing feature files if possible.

### Step2: Create steps definition

* Generate steps.go for your case. Under directory test/e2e, run:

```bash
$ go run hack/codegen.go -dest-path=steps/<your-new-feature>  features/<your-new-feature>/<your-new-feature>.feature
```
  

* Edit generated code, implement all generated functions. If you reuse step description from other feature file, you can also reuse corresponding function from that `step.go` file in this step.

### Step3: Add Init function into e2e_test.go

* In e2e_test.go, add generated feature file and InitializeScenario function into map `features`.

```go
var (
	features = map[string]InitialFunc{
		"features/conformance/host_rules.feature":           {hostrules.InitializeScenario, nil},
    ...
	}
)

```

### Step4: Build and run



* Build e2e_test:

```bash
$ make build
```

* Run your case, using `--feature` to specify the feature file.

```bash
$ ./e2e_test --feature features/<your-new-feature>/<your-new-feature>.feature
```
> Before running, your testing environment must be ready.

* Run all cases
```bash
$ ./run.sh
```

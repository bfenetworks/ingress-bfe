# BFE ingress controller e2e test

The e2e test follows K8s project [ingress-controller-conformance](https://github.com/kubernetes-sigs/ingress-controller-conformance), and add more test cases for ingress-bfe special features.

## Running

In ingress-bfe project's top directory, execute:
``` 
$ make e2e-test
```
It would automatically start the whole testing with following procedures:

- Build bfe-ingress-controller docker image
- Prepare test environment, including setting up a local k8s cluster by [Kind](https://kind.sigs.k8s.io/), loading and applying the docker images, etc. The scripts used for preparing environment is located in [test/script](../script).
- Execute test cases by running [run.sh](./run.sh), which actually build and execute program e2e_test.

## Contributing

We encourage contributors write e2e test case for new feature needed to be merged into ingress-bfe.

The test code is based on BDD testing framework [godog](https://github.com/cucumber/godog), a testing framework of [cucumber](https://cucumber.io/). It uses [Gherkin Syntax]( https://cucumber.io/docs/gherkin/reference/) to describe test case.

Steps to add new test case as below:

### Step1: Create Gherkin feature

* Create feature file to describe your test case. 

All existing feature files is under directory [test/e2e/features](./features). Please put your feature file into proper directory.
  > Reuse existing steps from existing feature files.

### Step2: Create steps definition

* Generate steps.go for your case. Under directory test/e2e, run:

```bash
$ go run hack/codegen.go -dest-path=steps/<dest-dir>  features/<xxx>.feature
```
  

* Edit generated code, implement all generated functions. If you reuse step in other feature file, you also can reuse the logic from exist steps files.

### Step3: Add step into e2e_test.go

* In e2e_test.go, add generated feature file and InitializeScenario function into map `features`.

```go
var (
	features = map[string]func(*godog.ScenarioContext){
		"features/conformance/host_rules.feature":           hostrules.InitializeScenario,
    ...
	}
)

```

### Step4: Build and run case

* Build e2e_test
```bash
$ make build
$ ./e2e_test --feature features/<xxx>.feature
```

After your test case pass, commit and push the code.

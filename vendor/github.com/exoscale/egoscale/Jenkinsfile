@Library('jenkins-pipeline') _

node {
  cleanWs()

  repo = 'exoscale/egoscale'

  try {
    dir('src') {
      stage('SCM') {
        checkout scm
      }
      stage('gofmt') {
        gofmt(repo, "cs", "exo")
      }
      updateGithubCommitStatus('PENDING', "${env.WORKSPACE}/src")
      stage('Build') {
        parallel (
          "golint": {
            golint(repo, "cmd/cs/...", "cmd/exo/...", "generate")
          },
          "go test": {
            test(repo)
          },
          "go install": {
            build(repo, "cs", "exo")
          },
        )
      }
    }
  } catch (err) {
    currentBuild.result = 'FAILURE'
    throw err
  } finally {
    if (!currentBuild.result) {
      currentBuild.result = 'SUCCESS'
    }
    updateGithubCommitStatus(currentBuild.result, "${env.WORKSPACE}/src")
    cleanWs cleanWhenFailure: false
  }
}

def gofmt(repo, ...bins) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh 'test `gofmt -s -d -e . | tee -a /dev/fd/2 | wc -l` -eq 0'
      // let's not gofmt the dependencies
      sh "cd /go/src/github.com/${repo} && dep ensure -v -vendor-only"
      for (bin in bins) {
        sh "cd /go/src/github.com/${repo}/cmd/${bin} && dep ensure -v -vendor-only"
      }
    }
  }
}

def golint(repo, ...extras) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "golint -set_exit_status -min_confidence 0.6 \$(go list github.com/${repo}/... | grep -v /vendor/)"
      sh "go vet `go list github.com/${repo}/... | grep -v /vendor/`"
      sh "cd /go/src/github.com/${repo} && gometalinter ."
      for (extra in extras) {
        sh "cd /go/src/github.com/${repo} && gometalinter ./${extra}"
      }
    }
  }
}

def test(repo) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "cd /go/src/github.com/${repo} && go test"
    }
  }
}

def build(repo, ...bins) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.10')
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/exoscale/egoscale") {
      for (bin in bins) {
        sh "go install github.com/${repo}/cmd/${bin}"
        sh "test -e /go/bin/${bin}"
      }
    }
  }
}

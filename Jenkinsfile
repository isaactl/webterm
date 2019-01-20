pipeline {
  agent {
    docker {
      image 'golang:1.10-alpine'
      args '-v $WORKSPACE:/go/src/github.com/isaactl/webterm -w /go/src/github.com/isaactl/webterm'
    }

  }
  stages {
    stage('build') {
      steps {
        sh 'go version && cd /go/src/github.com/isaactl/webterm  && pwd &&  go build'
      }
    }
  }
}
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
        sh 'cd /go/src/github.com/isaactl/webterm && ls -al  && env'
        sh 'go version && pwd && go build'
      }
    }
  }
}
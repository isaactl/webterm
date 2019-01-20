pipeline {
  agent any
  stages {
    stage('build exec') {
      agent {
        docker {
          image 'golang:1.10-alpine'
          args '-v $WORKSPACE:/go/src/github.com/isaactl/webterm'
        }

      }
      steps {
        sh 'go version && cd /go/src/github.com/isaactl/webterm  && pwd &&  go build -o ./bin/webterm '
      }
    }
    stage('Build image') {
      steps {
        sh 'ls -al && docker build -t webterm .'
      }
    }
  }
}
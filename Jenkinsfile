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
        sh 'go version && cd /go/src/github.com/isaactl/webterm  && pwd &&  go build -o ./bin/webterm && ls -al '
      }
    }
    stage('Build image') {
      agent any
      steps {
        sh 'pwd && ls -al && docker build -t webterm .'
      }
    }
  }
}
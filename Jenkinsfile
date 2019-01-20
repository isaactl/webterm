pipeline {
  agent none
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
        stash(name: 'store exec file', includes: 'bin/webterm', useDefaultExcludes: true)
      }
    }
    stage('Build image') {
      agent any
      steps {
        unstash 'store exec file'
        sh 'pwd && ls -al && docker build -t webterm .'
      }
    }
  }
}
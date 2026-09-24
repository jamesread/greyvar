node() {
	stage("Build") {
		deleteDir()

		checkout scm

		sh "mvn package"

		archive "target/*.jar"
	}
}

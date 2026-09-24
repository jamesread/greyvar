#!groovy

properties (
    [
         buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')),
         [$class: 'CopyArtifactPermissionProperty', projectNames: '*'],
         pipelineTriggers([[$class: 'PeriodicFolderTrigger', interval: '1d']])
    ]
)

node("fedora") {
	stage("Prep") {
		deleteDir();

		checkout scm
		sh "buildid -n"
	}

	stage("Code Analysis") {
		sh "flawfinder src/* > rpt/flawfinder.txt"
		sh "cppcheck src/* > rpt/cppcheck.txt"
	}

	stage("Build-lin") {
		sh "rm -rf build CMakeFiles CMakeCache.txt "
		sh 'cmake .'
		sh 'make'

	}

	stage("Build-win") {
		sh "rm -rf build CMakeFiles CMakeCache.txt "
		sh 'cmake -DCMAKE_TOOLCHAIN_FILE=windows.toolchain .'
		sh 'make'
	}

	stage("Package") {
		sh "ant package"

		archive 'pkg/**'

		stash includes: 'pkg/**', name: 'binaries'
	}
}

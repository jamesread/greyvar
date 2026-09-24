#include <boleas/boleas.hpp>

#include "gui/screens/Screens.hpp"

#include <boleas/GameState.hpp>
#include <boleas/gui/Gui.hpp>
#include <boleas/cvars.hpp>
#include <boleas/net/NetClient.hpp>

#include <cstdlib>
#include <ctime>
#include <iostream>

using std::cout;
using std::endl;

void hookMainLoop() {
	if (NetClient::get().isReady()) {
		NetClient::get().sendRecvFrame();
	}
}

void hookWindowReady() {
	cout << "Hook window ready" << endl;
	if (cvarIsset("server")) {
		cout << "Hook window ready2" << endl;
		NetClient::get().connect();
		NetClient::get().playerSetup(GameState::get().getFirstLocalPlayer());
		Gui::get().scene = PLAY;
	}
}

void* processServerFrames(void*) {
	while (true) {
		NetClient::get().processServerFrames();
	}

	boleasQuitEngine();
	pthread_exit(0);
}

void setupNetworkThread() {
	pthread_t serverRecvThread;
	pthread_create(&serverRecvThread, NULL, processServerFrames, NULL);
	pthread_setname_np(serverRecvThread, "serverRecv");
}

int mainGreyvarCore(int argc, char* argv[]) {
	cout << "Greyvar (core) " << endl << "--------------" << endl;

	setupScreens();

	setupNetworkThread();

	boleasHookInput = &hookMainLoop;
	boleasHookWindowReady = &hookWindowReady; 

	Gui::get().setScreen("main");

	boleasPrintVersion();
	boleasStartEngine(argc, argv);

	cout << "Everything has quit. Bye! " << endl;

	return 0;
}

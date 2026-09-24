#include "ScreenServerBrowser.hpp"

#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/components/Button.hpp>
#include <boleas/gui/components/Menu.hpp>
#include <boleas/gui/Gui.hpp>
#include <boleas/cvars.hpp>
#include <boleas/net/NetClient.hpp>

#include <vector>

class Server {
	public:
		Server(string address, int ping) {
			this->address = address;
			this->ping = ping;
		}

		string address;
		int ping;
};

vector<Server*> servers; 

void ScreenServerBrowser::setupComponents() {
	auto cons = new LayoutConstraints();

	servers.push_back(new Server("localhost", 123));

	auto lbl = new Label("Greyvar \u00BB Servers", 48);
	cons->rowWeight = 0;
	this->add(lbl, cons);

	cons->row++;
	cons->rowWeight = 0;
	this->add(new Label("^3Address"), cons);

	cons->col++;
	this->add(new Label("^3Ping"), cons);

	cons->col++;
	this->add(new Label("^3Details"), cons);

	for (auto server : servers) {
		cons->row++;
		cons->col = 0;
		this->add(new Label(server->address, 24, false, false, false), cons);

		cons->col++;
		this->add(new Label(to_string(server->ping), 24, false, false, false), cons);

		cons->col++;
		this->add(new Label("...", 24, false, false, false), cons);

		cons->col++;
		auto button = new Button("Connect", []() { 
			//cvarSet("server", server->address);
			Gui::get().setScreen("playerSetup");
		});
		this->add(button, cons);
	}

	auto btnBack = new Button("Back", []() {
		Gui::get().setScreen("main");
	});

	cons->row++;
	cons->rowWeight = 0;
	cons->col = 0;
	this->add(btnBack, cons);
}


void ScreenServerBrowser::executeCurrentItem() {
	NetClient::get().connect();
}

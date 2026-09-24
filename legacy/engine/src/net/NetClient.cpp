#include "net/NetClient.hpp"

#include "util.hpp"
#include "gui/Gui.hpp"
#include "LocalPlayer.hpp"
#include <unistd.h>

void NetClient::connect() {
	ConnectionRequest req; 
	ConnectionResponse res;
	ClientContext ctx;

	this->uuid = uuidGen();

	//ctx.AddMetadata("client-uuid", this->uuid);

	stub_->Connect(&ctx, req, &res);

	Gui::get().addMessage("connected to server");

	cout << "server version: " << res.serverversion() << ", client-uuid: " << uuid << endl;
}

void NetClient::playerSetup(LocalPlayer* lp) {
	NewPlayer np; 
	np.set_playerid(lp->playerId);
	NoResponse ignoreme;
	np.set_username("untitled.player");

	ClientContext ctx;
	
	//ctx.AddMetadata("client-uuid", this->uuid);

	stub_->PlayerSetup(&ctx, np, &ignoreme);
}

void NetClient::sendRecvFrame() {
	ServerFrameResponse res;
	ClientContext ctx;
	
	//ctx.AddMetadata("client-uuid", this->uuid);

	auto s = stub_->GetServerFrame(&ctx, *this->nextFrameToSend, &res);

	this->nextFrameToSend = new ClientRequests();
//	this->nextFrameToSend->set_clientid(GameState::get()->getFirstLocalPlayer()->playerId);

	if (!s.ok()) {
		cout << "omg, srv frame fail: " << s.error_message() << endl;
	} else {
		cout << "frame push" << endl;
		//serverFrameBuffer.push(res);
	}
}

void NetClient::processServerFrames() {
	while (this->hasFrames()) {
		auto frame = this->serverFrameBuffer.front();
		this->serverFrameBuffer.pop();

		processServerFrame(frame);
	}
}

bool NetClient::hasFrames() {
	//cout << "hasFrames " << this->serverFrameBuffer.size() << endl;
	return !this->serverFrameBuffer.empty();
}

NetClient& NetClient::get() {
	static NetClient instance(grpc::CreateChannel("localhost:2000", grpc::InsecureChannelCredentials()));
	
	return instance;
}


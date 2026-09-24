#pragma once

#include <grpc++/grpc++.h>
#include "proto/server_interface.grpc.pb.h"
#include "../LocalPlayer.hpp" 

#include <queue>

using greyvarproto::ServerInterface;
using greyvarproto::ConnectionRequest;
using greyvarproto::NewPlayer;
using greyvarproto::MoveRequest; 
using greyvarproto::NoResponse;
using greyvarproto::ConnectionResponse;
using greyvarproto::ClientRequests;
using greyvarproto::ServerFrameResponse;
using grpc::Channel;
using grpc::ClientContext;

using std::cout;
using std::endl;

void processServerFrame(ServerFrameResponse sframe);
void processServerFrameGrid(greyvarproto::Grid grid);
void processServerFrameEntitySpawns(greyvarproto::EntitySpawn spawn);

class NetClient {
	public:
		NetClient(NetClient const&) = delete;
		void operator=(NetClient const&) = delete;

		static NetClient& get();

		ClientRequests* nextFrameToSend = new ClientRequests(); 

		void connect();
		void playerSetup(LocalPlayer* lp);

		ClientRequests* getNextFrameToSend() {
			return this->nextFrameToSend;
		}

		bool hasFrames();

		void processServerFrames();

		bool isReady() {
			auto state = this->channel->GetState(false);

			return state == GRPC_CHANNEL_READY;
		}

		void sendRecvFrame();

	private:
		std::string uuid;

		NetClient(std::shared_ptr<Channel> channel) : stub_(ServerInterface::NewStub(channel)) {
			this->channel = channel;

			cout << "NetClient constructed" << endl;
			cout << "cons size " << this->serverFrameBuffer.size() << endl;
		}
	
		std::queue<ServerFrameResponse> serverFrameBuffer;

		std::shared_ptr<Channel> channel;
		std::unique_ptr<ServerInterface::Stub> stub_;
};


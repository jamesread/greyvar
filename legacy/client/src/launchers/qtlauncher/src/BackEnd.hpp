#pragma once

#include <QObject>
#include <curl/curl.h>

#include "Updater.hpp"

class BackEnd : public QObject {
	Q_OBJECT
	Q_PROPERTY(QString status READ getStatus NOTIFY statusChanged)
	Q_PROPERTY(bool canPlay READ canPlay)

	public: 
		BackEnd() {}
		~BackEnd() {}

		Q_INVOKABLE void checkStatus() {
			if (this->isAlreadyCheckingStatus) {
				return;
			}

			this->isAlreadyCheckingStatus = true;

			this->updateStatus("Checking...", false);

			
			int id = 1;
			pthread_t inc_x_thread;
			pthread_create(&inc_x_thread, NULL, runUpdateCheck, &id);
			
			this->isAlreadyCheckingStatus = false;
		}

		bool canPlay() {
			return this->playable;
		}
			
		QString getStatus() {
			return this->status;
		}

	signals:
		void statusChanged();
	
	private:
		pthread_t thread;

		void updateStatus(QString status, const bool playable) {
			this->status = status;
			this->playable = playable;

			this->statusChanged();
		}

		QString status = "not yet checked";

		bool playable = false;
		bool isAlreadyCheckingStatus = false;
};



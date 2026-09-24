#include "BackEnd.hpp"

#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQuickView>
#include <QObject>
#include <iostream>

using std::cout;
using std::endl;

int main(int argc, char** argv) {
	cout << "Greyvar launcher" << endl;

	QGuiApplication app(argc, argv);
	QQmlApplicationEngine engine;

	qmlRegisterType<BackEnd>("io.github.greyvar.launcher", 1, 0, "BackEnd");
	
	engine.load(QUrl("qrc:/window.qml"));

//	QObject::connect(view.rootObject(), SIGNAL(qmlSignal(QString)), &thing, SLOT(cppSlot(QString)));

	return app.exec();
}

package greyvarEditor.utils;

import java.io.File;

import javax.swing.JFileChooser;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.Logger;

import org.w3c.dom.Document;
import org.w3c.dom.Element;

import greyvarEditor.ui.windows.WindowProgress;

public class TextureProertiesReader implements Runnable {

	private WindowProgress progress;
	private File selectedFile; 

	public TextureProertiesReader() { 
		JFileChooser jfc = new JFileChooser(); 
		jfc.setDialogTitle("Load XML props");
		jfc.setCurrentDirectory(new File(Configuration.getS("paths.xmlProps")));
		jfc.showOpenDialog(null);

		if (jfc.getSelectedFile() != null) {
			this.selectedFile = jfc.getSelectedFile();
			new Thread(this).start();
		}

		Configuration.set("paths.xmlProps", jfc.getCurrentDirectory().getAbsolutePath());
	}

	@Override
	public void run() {
		this.progress = new WindowProgress();

		try {
			DocumentBuilder db = DocumentBuilderFactory.newInstance().newDocumentBuilder();
			Document d = db.parse(this.selectedFile);

			Element root = (Element) d.getElementsByTagName("root").item(0);

			Logger.messageDebug("root: " + root);
		} catch (Exception e) {
			e.printStackTrace();
		}

	}
}

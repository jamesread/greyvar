package greyvarEditor;


import java.io.File;
import java.nio.file.Path;

import javax.swing.UIManager;

import greyvarEditor.files.GridFile;
import greyvarEditor.files.GridFileLoader;
import greyvarEditor.files.WorldFileHandler;
import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.WindowEditorWorld;
import greyvarEditor.ui.windows.WindowMain;
import jwrCommonsJava.Configuration;
import jwrCommonsJava.Logger;

public class Main {
	public static void main(String[] args) throws Exception {
		try {
			UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
		} catch (Exception e) {

			e.printStackTrace();
		}
	
		Configuration.set("sysAppName", "greyvarEditor");
		Configuration.parseInternalConfiguration();

		if (!Configuration.isUserConfigExistering()) {
			Configuration.createUserConfig();
		}

		Configuration.parseUserConfiguration();
		
		Logger.messageDebug("Textures (paths.textureDirectory): " + GameResources.dirTextures);

		TextureCache.instanceTiles.getDefault();

		WindowMain.getInstance().setVisible(true);
		
		if (args.length > 0) {
			File f = new File(args[0]);
			if (f.exists()) {
				if (f.getName().equals("world.yml")) {    
					WindowMain.getInstance().addFrame(new WindowEditorWorld(f));
				} else if (f.getName().endsWith("grid")) { 
					GridFile gf = GridFileLoader.load(f);
					WindowMain.getInstance().addFrame(new WindowEditorGrid(gf));	
				}
			}
		}
	} 
}

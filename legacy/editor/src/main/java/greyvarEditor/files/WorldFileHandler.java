package greyvarEditor.files;

import java.io.File;
import java.util.Vector;

import greyvarEditor.dataModel.World;
import greyvarEditor.triggers.Trigger;
import greyvarEditor.utils.YamlFile;
import jwrCommonsJava.Logger; 

public class WorldFileHandler {
	public static World load(File f) { 
		try {
			World w = YamlFile.read(f, World.class);
			w.file = f; 
			
			loadGrids(w, new File(f.getParentFile(), "grids"));
			 
			return w;
		} catch (Exception e) { 
			Logger.messageException(e, "Could not load world file"); 
		}
		 
		return null;
	}
	 
	private static void loadGrids(World w, File worldDirectory) {
		for (File f : worldDirectory.listFiles()) {  
			w.gridFiles.add(f);  
		}
	}

	public static void save(World world) {   
		YamlFile.write(world.file, world);  
	}
}

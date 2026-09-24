package greyvarEditor;

import java.io.File;

import jwrCommonsJava.Configuration;

public class GameResources {
	public static File dirServerData = new File(Configuration.getS("paths.serverDatDirectory"));
	public static File dirTextures = new File(Configuration.getS("paths.textureDirectory"));
	
	public static File dirServerDataWorlds = new File(dirServerData, "worlds");
	 
	public static EntityDatabase entityDatabase = new EntityDatabase(new File(dirServerData, "/entdefs"));
}
 
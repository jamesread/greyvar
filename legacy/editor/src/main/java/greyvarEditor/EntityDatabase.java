package greyvarEditor;

import java.io.File;
import java.io.FilenameFilter;
import java.io.IOException;
import java.util.HashMap;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;

import greyvarEditor.dataModel.EntityDefinition;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import jwrCommonsJava.Logger;

public class EntityDatabase {
	public HashMap <String, EntityDefinition> entdefs = new HashMap<>();
	private File baseDir;

	public EntityDatabase(File directory) {
		this.baseDir = directory;
		
		this.reloadEntdefs();		 
	}
	
	public void reloadEntdefs() {
		this.entdefs.clear();
		
		if (this.baseDir.exists()) {
			for (File f: this.baseDir.listFiles((File dir, String filename) -> { return filename.endsWith(".yml"); })) {
				this.loadEntdef(f);
			}
		} 
	}
	
	public void loadEntdef(File f) {
		ObjectMapper om = new ObjectMapper(new YAMLFactory());
		 
		try {
			EntityDefinition entdef = om.readValue(f, EntityDefinition.class);
			 
			this.entdefs.put(entdef.title, entdef);
			  
			Logger.messageNormal("Loaded entdef: " + entdef.title + ", initial state = " + entdef.initialState);
		} catch (IOException e) {
			e.printStackTrace();
		}
	}
 
	public Texture getEntityTexture(String definition) {
		EntityDefinition entdef = this.entdefs.get(definition);
		
		if (entdef == null) {
			return TextureCache.instanceEntities.getDefault();	
		} else {
			return TextureCache.instanceEntities.getTex(entdef.getEditorInitialState().tex); 
		} 
	}

	public int size() {
		return entdefs.size(); 
	}

}

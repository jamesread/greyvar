package greyvarEditor.dataModel;

import greyvarEditor.GameResources;
import greyvarEditor.TextureCache;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import jwrCommonsJava.Logger;

public class EntityInstance {
	public int id;  
	public String definition;

	public transient EntityDefinition definitionLink;
	public transient Texture editorTexture;
	
	public EntityInstance() {
		
	}
	
	public void setDefinition(String def) {
		this.definition = def;
		
		this.definitionLink = GameResources.entityDatabase.entdefs.get(this.definition);
		
		if (this.definitionLink == null) { 
			Logger.messageWarning("Could not find entdef: " + this.definition);
		} else {
			this.editorTexture = TextureCache.instanceEntities.getTex(this.definitionLink.getEditorInitialState().tex);
		}
	}  
 
	public EntityInstance(String definition, int id) {
		this.id = id; 
		 
		this.setDefinition(definition); 
	}
}

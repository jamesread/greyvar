package greyvarEditor.dataModel;

import java.beans.Transient;
import java.util.HashMap;

import jwrCommonsJava.Logger;

public class EntityDefinition {
	public String title; 
	public HashMap<String, EntityState> states = new HashMap<>();
	public String initialState; 
	public String texture;
	
	@Transient  
	public EntityState getEditorInitialState() {
		if (states.size() == 0) {
			throw new RuntimeException("Entity definition that has zero states!: " + this.title); 	
		} else { 
			if (states.containsKey(initialState)) {
				return states.get(initialState);
			} else { 
				Logger.messageWarning("Correct initial state not found, using the first state instead for ent:" + this.title);
				return states.values().iterator().next();
			}
		}
	}
}

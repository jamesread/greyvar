package greyvarEditor.triggers.eventFragments;

import greyvarEditor.triggers.Action;
import greyvarEditor.triggers.arguments.ArgumentEntityId;
import greyvarEditor.triggers.arguments.ArgumentTileCoordinates;

public class ActionEntityMove extends Action {
	private ArgumentEntityId argEntId = new ArgumentEntityId(-1);
	private ArgumentTileCoordinates argPos = new ArgumentTileCoordinates(0, 0);
	
	public ActionEntityMove() { 
		this.arguments.put("entity", argEntId);
		this.arguments.put("pos", argPos);
	} 
	
	@Override
	public String getDescription() {
		return "move entity "; 
	}  
	
	public String getArgumentDescriptionPrefix(String key) {
		if (key.equals("pos")) {
			return "to"; 
		} 
		  
		return "";
	}
	 
	public ActionEntityMove set(int entityId, int x, int y) {
		this.argEntId.id = entityId;
		this.argPos.x = x;
		this.argPos.y = y; 
		
		return this; 		
	}
} 

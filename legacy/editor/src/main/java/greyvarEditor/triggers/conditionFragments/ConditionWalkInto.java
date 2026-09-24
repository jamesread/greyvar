package greyvarEditor.triggers.conditionFragments;

import greyvarEditor.triggers.Condition;
import greyvarEditor.triggers.arguments.ArgumentTileCoordinates;

public class ConditionWalkInto extends Condition {
	private ArgumentTileCoordinates pos = new ArgumentTileCoordinates(0, 0);
	 
	public String getDescription() {
		return "player walks into";  
	}
	 
	public ConditionWalkInto() {   
		this.arguments.put("pos", pos);
	}
}

package greyvarEditor.triggers.conditionFragments;

import greyvarEditor.triggers.Condition;
import greyvarEditor.triggers.arguments.ArgumentTileCoordinates;

public class ConditionStepOn extends Condition { 
	private ArgumentTileCoordinates pos = new ArgumentTileCoordinates(0, 0);
			
	public ConditionStepOn() {   
		this.arguments.put("pos", pos);
	}
	
	public String getDescription() {
		return "player steps on tile";   
	}
 
	public ConditionStepOn pos(int x, int y) {
		this.pos.x = x;
		this.pos.y = y;
		 
		return this;
	}
}

package greyvarEditor.triggers;

import java.util.Vector;

import greyvarEditor.triggers.arguments.ArgumentEntityId;
import greyvarEditor.triggers.arguments.ArgumentTileCoordinates;
import greyvarEditor.triggers.conditionFragments.ConditionStepOn;
import greyvarEditor.triggers.eventFragments.ActionEntityMove;
import greyvarEditor.triggers.eventFragments.ActionMessage;

public class TriggerTypeRegistry {
	public static Vector<Class<? extends Condition>> conditions = new Vector<>();
	public static Vector<Class<? extends Action>> actions = new Vector<>(); 
	
	public static Vector<Class<? extends FragmentArgument>> fragmentArguments = new Vector<>();
	
	static {
		conditions.add(ConditionStepOn.class);
		
		actions.add(ActionEntityMove.class);
		actions.add(ActionMessage.class);
		 
		fragmentArguments.add(ArgumentEntityId.class); 
		fragmentArguments.add(ArgumentTileCoordinates.class);
	}
}

package greyvarEditorTests.gridLoading;

import greyvarEditor.triggers.Statement;
import greyvarEditor.triggers.Trigger;
import greyvarEditor.triggers.conditionFragments.ConditionStepOn;
import greyvarEditor.triggers.eventFragments.ActionEntityMove;
import greyvarEditor.triggers.eventFragments.ActionMessage;
import greyvarEditor.ui.windows.WindowTriggerFragmentEditor;

public class TestRenderStatement {
	public static void main(String[] args) {
		testRenderStatement();
	}
	
	public static void testRenderStatement() {  
		Trigger triggerOpenDoor = new Trigger();
		
		triggerOpenDoor.conditions.add(new ConditionStepOn().pos(3, 9)); 
		triggerOpenDoor.actions.add(new ActionEntityMove().set(1, 7,  12));   
		triggerOpenDoor.actions.add(new ActionMessage().set("Door Opened!")); 
		
		new WindowTriggerFragmentEditor(new ConditionStepOn().pos(3,4));  
	}
}

package greyvarEditor.triggers.eventFragments;

import greyvarEditor.triggers.Action;
import greyvarEditor.triggers.Fragment;
import greyvarEditor.triggers.arguments.ArgumentString;
 
public class ActionMessage extends Action {
	private ArgumentString argMsg = new ArgumentString();

	public ActionMessage() {
		this.arguments.put("message", argMsg); 
	}
	
	@Override
	public String getDescription() {
		return "show the message"; 
	} 
	
	public ActionMessage set(String message) {
		this.argMsg.value = message;
		 
		return this; 
	}

}

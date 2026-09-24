package greyvarEditor.triggers;

import java.util.Vector;

public class Trigger {
	public String title = "unamed trigger";  
	public Vector<Condition> conditions = new Vector<>();
	public Vector<Action> actions = new Vector<>();   
	
	@Override
	public String toString() {
		return title; 
	}
}

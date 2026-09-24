package greyvarEditor.triggers;

import java.util.Vector;

import greyvarEditor.triggers.conditionFragments.ConditionPlayerLeave;
import greyvarEditor.triggers.conditionFragments.ConditionPlayerSpawn;
import greyvarEditor.triggers.conditionFragments.ConditionStepOn;
import greyvarEditor.triggers.conditionFragments.ConditionWalkInto;
import greyvarEditor.triggers.eventFragments.ActionDisable;
import greyvarEditor.triggers.eventFragments.ActionEntityMove;
import greyvarEditor.triggers.eventFragments.ActionMessage;

public class LibraryOfFragments {
	private Vector<Class<? extends Condition>> conditions = new Vector<>();
	private Vector<Class<? extends Action>> actions = new Vector<>();
	
	public static LibraryOfFragments instance = new LibraryOfFragments();
	 
	private LibraryOfFragments() {
		this.conditions.add(ConditionPlayerLeave.class);
		this.conditions.add(ConditionPlayerSpawn.class);
		this.conditions.add(ConditionStepOn.class); 
		this.conditions.add(ConditionWalkInto.class);
		
		this.actions.add(ActionEntityMove.class); 
		this.actions.add(ActionMessage.class); 
		this.actions.add(ActionDisable.class); 
	}
	 
	private Class<? extends Action> getActionByName(String name) {
		for (Class<? extends Action> c : this.actions) {
			if (c.getSimpleName().equals(name)) {
				return c;
			}
		} 
		
		return null;
	}
	
	private Class<? extends Condition> getConditionByName(String name) {
		for (Class<? extends Condition> c : this.conditions) {
			if (c.getSimpleName().equals(name)) {
				return c;
			}
		}
		
		return null;
	}
	
	public Action createAction(String name) {
		try {
			Action a = this.getActionByName(name).newInstance();
			
			return a;
		} catch (InstantiationException | IllegalAccessException e) {
			e.printStackTrace();
		}  
		
		return null;
	}
	
	public Condition createCondition(String name) {
		try {
			Condition c = this.getConditionByName(name).newInstance();
			
			return c;
		} catch (InstantiationException | IllegalAccessException e) {
			e.printStackTrace();
		} 
		
		return null;
	}
	 
	public Vector<String> getConditions() {
		Vector<String> names = new Vector<>(); 
		
		for (Class<? extends Condition> f : this.conditions) {
			names.add(f.getSimpleName());
		} 
		
		return names;
	} 
	
	public Vector<String> getActions() {
		Vector<String> names = new Vector<>();
		
		for (Class<? extends Action> f : this.actions) {
			names.add(f.getSimpleName());
		}
		
		return names; 
	}
}

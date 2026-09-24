package greyvarEditor.triggers;

import java.beans.Transient;
import java.util.HashMap;
import java.util.Map;
import java.util.SortedMap;
import java.util.TreeMap;
import java.util.Vector;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonSubTypes.Type;

import greyvarEditor.triggers.conditionFragments.ConditionPlayerLeave;
import greyvarEditor.triggers.conditionFragments.ConditionPlayerSpawn;
import greyvarEditor.triggers.conditionFragments.ConditionStepOn;
import greyvarEditor.triggers.conditionFragments.ConditionWalkInto;
import greyvarEditor.triggers.eventFragments.ActionDisable;
import greyvarEditor.triggers.eventFragments.ActionEntityMove;
import greyvarEditor.triggers.eventFragments.ActionMessage;
import greyvarEditor.triggers.eventFragments.ActionSetEntityState;

import com.fasterxml.jackson.annotation.JsonTypeInfo;
 
@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes({
	@Type(value = ConditionPlayerLeave.class, name = "conditionPlayerLeave"),
	@Type(value = ConditionPlayerSpawn.class, name = "conditionPlayerSpawn"),
	@Type(value = ConditionStepOn.class, name = "conditionPlayerStepOn"),
	@Type(value = ConditionWalkInto.class, name = "conditionPlayerWalkInto"), 
	@Type(value = ActionDisable.class, name = "actionDisable"),
	@Type(value = ActionEntityMove.class, name = "actionEntityMove"),
	@Type(value = ActionMessage.class, name = "actionMessage"), 
	@Type(value = ActionSetEntityState.class, name = "actionSetEntityState")
})
public class Fragment { 
	protected Map<String, FragmentArgument> arguments = new TreeMap<>(); 
	
	@Transient
	public String getDescription() {
		return "<generic fragment>"; 
	}
	 
	@Transient
	public String getArgumentDescriptionPrefix(String argumentName) {
		return "";
	}
	 
	public Map<String, FragmentArgument> getArguments() {
		return this.arguments; 
	}

	@Transient
	public FragmentArgument getArgument(String argName) {
		return this.arguments.get(argName); 
	}
	
	@Override
	public String toString() {
		String ret = "";
		
		ret += this.getDescription();
		   
		for (FragmentArgument fa : this.arguments.values()) {
			ret += " "; 
			ret += fa.toString();
		}  
				
		return ret;
	}
}  
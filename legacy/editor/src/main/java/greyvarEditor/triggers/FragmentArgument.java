package greyvarEditor.triggers;

import java.beans.Transient;

import javax.swing.JComponent;
import javax.swing.JLabel;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;
import com.fasterxml.jackson.annotation.JsonSubTypes.Type;

import greyvarEditor.triggers.arguments.ArgumentEntityId;
import greyvarEditor.triggers.arguments.ArgumentString;
import greyvarEditor.triggers.arguments.ArgumentTileCoordinates;
import greyvarEditor.triggers.conditionFragments.ConditionPlayerLeave;
import greyvarEditor.triggers.conditionFragments.ConditionPlayerSpawn;
import greyvarEditor.triggers.conditionFragments.ConditionStepOn;
import greyvarEditor.triggers.conditionFragments.ConditionWalkInto;
import greyvarEditor.triggers.eventFragments.ActionDisable;
import greyvarEditor.triggers.eventFragments.ActionEntityMove;
import greyvarEditor.triggers.eventFragments.ActionMessage;
import greyvarEditor.triggers.eventFragments.ActionSetEntityState;
    
@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes({
	@Type(value = ArgumentEntityId.class, name = "entId"), 
	@Type(value = ArgumentString.class, name = "string"),   
	@Type(value = ArgumentTileCoordinates.class, name = "tileCoords")  
}) 
public abstract class FragmentArgument {
	public FragmentArgument() {
	}
	
	@Override 
	public String toString() {
		return this.getClass().getSimpleName(); 
	}
 
	@Transient
	public JComponent getEditor() {
		return new JLabel("(no editor available)");
	}
	 
	public abstract void setFromEditor();
}

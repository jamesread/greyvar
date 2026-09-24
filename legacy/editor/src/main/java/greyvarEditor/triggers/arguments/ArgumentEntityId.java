package greyvarEditor.triggers.arguments;

import java.awt.FlowLayout;

import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.JPanel;

import greyvarEditor.triggers.FragmentArgument; 

public class ArgumentEntityId extends FragmentArgument {
	public int id;
	public String grid;
	
	private class EntitySelector extends JPanel {
		private JComboBox<String> cboGrids = new JComboBox<>();
		private JComboBox<String> cboEntities = new JComboBox<>();
		
		public EntitySelector() {
			this.setLayout(new FlowLayout());
			
			this.add(new JLabel("Entity")); 
			this.add(cboEntities);
			
			this.add(new JLabel("Grid"));
			this.add(cboGrids);
		} 
	}
	
	private EntitySelector editor = new EntitySelector();
	
	public ArgumentEntityId(int entityId) {
		this.id = entityId; 
	} 
	
	@Override
	public JComponent getEditor() {
		return editor; 
	}

	@Override
	public void setFromEditor() {
		
	} 
	
	@Override
	public String toString() { 
		return "entity {" + id + "} on grid {" + grid + "}";
	}
}

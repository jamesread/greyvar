package greyvarEditor.ui.windows;

import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.Insets;

import javax.swing.JButton;
import javax.swing.JDialog;
import javax.swing.JFrame;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.ListModel;
import javax.swing.ListSelectionModel;
import javax.swing.event.ListDataListener;
import javax.swing.event.ListSelectionEvent;
import javax.swing.event.ListSelectionListener;

import greyvarEditor.dataModel.World;
import greyvarEditor.files.WorldFileHandler;
import greyvarEditor.triggers.Trigger;
import greyvarEditor.ui.components.ComponentCrudList;
import greyvarEditor.utils.YamlFile;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowTriggerList extends JFrame { 
	private ComponentCrudList<Trigger> triggerList;  
	
	private final World world; 
	
	public WindowTriggerList(JFrame parent, World world) {
		this.world = world;
		 
		this.triggerList = new ComponentCrudList<Trigger>(world.triggers) {
			
			@Override
			public void onEdit(Trigger t) {
				new WindowTriggerEdit(WindowTriggerList.this, t).setVisible(true);	
			}
			
			@Override
			public void onDelete(Trigger t) {
				WindowTriggerList.this.world.triggers.remove(t);
			} 
			
			@Override
			public void onCreate() {
				Trigger t = new Trigger();  
				WindowTriggerList.this.world.triggers.add(t); 
				  
				new WindowTriggerEdit(WindowTriggerList.this, t).setVisible(true);
			} 
		};
		
		this.setupComponents();  
		
		this.setTitle("Triggers for world: " + world.title);
		this.setBounds(100, 100, 640, 480);
		this.setLocationRelativeTo(null); 
		this.setVisible(true); 
	} 
	
	@Override
	public void setVisible(boolean b) {
		super.setVisible(b);
		
		if (!b) { 
			WorldFileHandler.save(world);
		}
	}
	 
	private void setupComponents() { 
		this.setIconImage(WindowMain.getInstance().getIconImage());
		
		GridBagConstraints gbc = new GridBagConstraints(0, 0, 1, 1, 1, 1, GridBagConstraints.CENTER, GridBagConstraints.BOTH, new Insets(6,6,6,6), 0, 0);  
		this.setLayout(new GridBagLayout());
		
		gbc.gridwidth = gbc.REMAINDER;
		this.add(triggerList, gbc);  
		 
	}
}

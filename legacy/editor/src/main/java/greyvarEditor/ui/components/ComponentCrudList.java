package greyvarEditor.ui.components;

import java.awt.Color;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.Insets;
import java.util.Vector;

import javax.swing.BorderFactory;
import javax.swing.JButton;
import javax.swing.JList;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.ListModel;
import javax.swing.ListSelectionModel;
import javax.swing.event.ListDataListener;
import javax.swing.event.ListSelectionEvent;
import javax.swing.event.ListSelectionListener;

import greyvarEditor.triggers.Trigger;
import jwrCommonsJava.Util;
import jwrCommonsJava.ui.JButtonWithAl;

public abstract class ComponentCrudList<T> extends AllEnabledOrDisabledPanel implements ListSelectionListener {
	private JList<T> lst;
	 
	private JButton btnCreate;
	private JButton btnEdit;
	private JButton btnDelete; 
	
	private final Vector<T> contents; 
	
	class Model implements ListModel<T> {

		@Override
		public int getSize() {
			return contents.size();
		}

		@Override
		public T getElementAt(int index) {
			return contents.elementAt(index);
		}

		@Override
		public void addListDataListener(ListDataListener l) { 
		}
 
		@Override
		public void removeListDataListener(ListDataListener l) {
		}
		
	}

	public ComponentCrudList(Vector<T> contents) {
		this.contents = contents;
		 
		this.setBorder(BorderFactory.createEmptyBorder());  
		
		lst = new JList<T>(new Model());
		lst.setSelectionMode(ListSelectionModel.SINGLE_SELECTION);
		lst.addListSelectionListener(this);
		
		this.setupComponents();
	}
	
	private void setupComponents() {
		GridBagConstraints gbc = Util.getNewGbc();
		gbc.insets = new Insets(0,0,0,0);    
		
		this.setLayout(new GridBagLayout());
		
		gbc.gridwidth = gbc.REMAINDER;
		gbc.weighty = 1;
		gbc.anchor = gbc.CENTER;
		gbc.fill = gbc.BOTH;
		this.add(new JScrollPane(lst), gbc);
		
		gbc.weighty = 0;
		gbc.weightx = 1;
		gbc.anchor = gbc.SOUTHEAST;
		gbc.fill = gbc.NONE;
		gbc.gridheight = 1; 
		gbc.gridwidth = 1;
		gbc.gridy++;
		gbc.gridx = 0;
		
		this.btnCreate = new JButtonWithAl("+") {
			
			@Override
			public void click() {
				onCreate();
			}
		};
		this.add(btnCreate, gbc);
		 
		gbc.weightx = 0; 
		gbc.gridx++;
		btnEdit = new JButtonWithAl("...") {
			
			@Override
			public void click() {
				T t = lst.getSelectedValue();
				
				if (t != null) {
					onEdit(t);	
				}
								
			}
		};
		this.add(btnEdit, gbc);
		
		gbc.gridx++;
		btnDelete = new JButtonWithAl("-") {
			
			@Override
			public void click() {
				T t = lst.getSelectedValue();
				
				if (t != null) {
					onDelete(t);	
				}
			}
		};
		this.add(btnDelete, gbc);
		
		updateButtons();
	}
	 
	public abstract void onCreate();
	public abstract void onDelete(T t);
	public abstract void onEdit(T t);
	public void onSelectionChanged() {}
	
	private void updateButtons() {
		T t = this.lst.getSelectedValue();
		
		if (t == null) {
			this.btnDelete.setEnabled(false);
			this.btnEdit.setEnabled(false);
		} else {
			this.btnDelete.setEnabled(true);
			this.btnEdit.setEnabled(true);
		}
	}
	
	T getSelected() {
		return lst.getSelectedValue();
	}
	 
	@Override
	public void valueChanged(ListSelectionEvent e) {
		updateButtons();
		onSelectionChanged();
	}

	public void refresh() {
		lst.updateUI(); 		
	}
	
}

package greyvarEditor.ui.windows;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.FlowLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.awt.event.MouseListener;

import javax.swing.JButton;
import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JDialog;
import javax.swing.JFrame;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SpringLayout;

import greyvarEditor.triggers.Action;
import greyvarEditor.triggers.Condition;
import greyvarEditor.triggers.Fragment;
import greyvarEditor.triggers.FragmentArgument;
import greyvarEditor.triggers.LibraryOfFragments;
import greyvarEditor.triggers.Statement;
import greyvarEditor.ui.components.ClickableFragmentLabel;
import greyvarEditor.ui.components.ClickableFragmentLabel.Listener;
import jwrCommonsJava.ui.JButtonWithAl;
  
@SuppressWarnings("serial") 
public class WindowTriggerFragmentEditor<T extends Fragment> extends JDialog implements Listener {
	private T fragment;
	
	private JButton btnClose;
	  
	private JPanel panSelectedComponent = new JPanel(); 
	private JPanel panStatement = new JPanel();
	
	private JButton btnSaveArg = new JButtonWithAl("Set") {
		@Override
		public void click() {
			setArg();
		}  
	}; 
	
	private enum FragmentEditType {
		UNKNOWN,
		CONDITION,
		ACTION
	}
	
	private FragmentEditType editMode = FragmentEditType.UNKNOWN;
	 
	public WindowTriggerFragmentEditor(T fragment ) { 
		this.setModalityType(ModalityType.APPLICATION_MODAL);
		this.setModalExclusionType(ModalExclusionType.APPLICATION_EXCLUDE); 
		 
		this.fragment = fragment;
		
		if (this.fragment instanceof Condition) {
			this.editMode = FragmentEditType.CONDITION;  
		} else if (this.fragment instanceof Action) {
			this.editMode = FragmentEditType.ACTION; 
		}
		
		this.setupComponents();
		
		this.renderStatement();
		
		this.setVisible(true);
	}
	
	private void renderStatement() {
		panStatement.removeAll();
		
		switch (editMode) {
		case CONDITION:
			panStatement.add(new JLabel("If"));
			  
			JLabel lblCondition = new ClickableFragmentLabel(this, fragment.getDescription(), cboConditions);
			lblCondition.setForeground(Color.MAGENTA); 
			panStatement.add(lblCondition);   
			
			break;  
		case ACTION: 
			panStatement.add(new JLabel("Then")); 
			
			JLabel lblAction = new ClickableFragmentLabel(this, fragment.getDescription(), cboActions);
			lblAction.setForeground(Color.ORANGE.darker()); 
			panStatement.add(lblAction);
			
			break; 
		}
		 
		for (String argName : this.fragment.getArguments().keySet()) {
			FragmentArgument fa = this.fragment.getArgument(argName);
			 
			if (!fragment.getArgumentDescriptionPrefix(argName).isEmpty()) { 
				panStatement.add(new JLabel(fragment.getArgumentDescriptionPrefix(argName)));
			} 
			 
			JLabel lbl = new ClickableFragmentLabel(this, fa.toString(), fa.getEditor(), argName);   
			lbl.setForeground(Color.BLUE);  
			panStatement.add(lbl);
		}
		
		switch (editMode) {
		case CONDITION: 
			panStatement.add(new JLabel("then ...")); 
		}
		
		panStatement.updateUI(); 
	} 
	
	private void closeWindow() {
		this.setVisible(false);
	}
	 
	private JComboBox<String> cboConditions = new JComboBox<>();
	private JComboBox<String> cboActions = new JComboBox<>(); 
	
	private void setupConditionSelect() {
		for (String cond : LibraryOfFragments.instance.getConditions()) {
			cboConditions.addItem(cond);	
		}  
		 
		cboConditions.addActionListener(new ActionListener() {
			@Override
			public void actionPerformed(ActionEvent e) {  
				String cond = cboConditions.getSelectedItem().toString();
				  
				fragment = (T) LibraryOfFragments.instance.createCondition(cond); 
				
				WindowTriggerFragmentEditor.this.renderStatement(); 
			}
		});
	}
	 
	private void setupActionSelect() {
		for (String action : LibraryOfFragments.instance.getActions()) {
			cboActions.addItem(action);
		}
		
		cboActions.addActionListener(new ActionListener() {
			
			@Override
			public void actionPerformed(ActionEvent e) { 
				String action = cboActions.getSelectedItem().toString();
				
				fragment = (T) LibraryOfFragments.instance.createAction(action);
				
				WindowTriggerFragmentEditor.this.renderStatement();
			}
		}); 
	}
	
	private void setupComponents() {
		this.setIconImage(WindowMain.getInstance().getIconImage());
		
		switch (editMode) {
		case CONDITION:
			this.setupConditionSelect();
			break;
		case ACTION:
			this.setupActionSelect(); 
		}
				
		this.setLayout(new GridBagLayout());
		
		GridBagConstraints gbc = jwrCommonsJava.Util.getNewGbc();
		
		
		gbc.fill = gbc.BOTH;
		gbc.weighty = 0;
		
		panStatement.setBackground(Color.WHITE);
		panStatement.setLayout(new FlowLayout());
		 
		this.add(new JScrollPane(panStatement), gbc);
		 
		gbc.weighty = 1;
		gbc.fill = gbc.NONE; 
		gbc.anchor = gbc.NORTH; 
		gbc.gridy++;   
		panSelectedComponent.setLayout(new BorderLayout()); 
		panSelectedComponent.add(new JLabel("(nothing selected)")); 
		this.add(panSelectedComponent, gbc);
		
		gbc.gridy++;  
		gbc.weighty = 0; 
		gbc.anchor = gbc.SOUTHEAST; 
		gbc.fill = gbc.NONE; 
		this.btnClose = new JButtonWithAl("Close") {
			
			@Override
			public void click() {
				closeWindow();
			}
		};
		add(btnClose, gbc); 
		
		this.setTitle("Trigger Fragment editor");
		this.doLayout();  
		this.setBounds(300, 300, 640, 240);
		this.setLocationRelativeTo(null); 
	}

	private FragmentArgument currentArgument; 
	
	@Override
	public void onFragmentClicked(JComponent comp, String argKey) {
		this.setArg();
		 
		this.currentArgument = this.fragment.getArgument(argKey);
		
		panSelectedComponent.removeAll(); 
		panSelectedComponent.add(comp, BorderLayout.CENTER);   
		panSelectedComponent.add(btnSaveArg, BorderLayout.EAST);  		
		panSelectedComponent.updateUI(); 
	}
	 
	public void setArg() {
		if (this.currentArgument != null) {
			this.currentArgument.setFromEditor();
		}
		 
		panSelectedComponent.removeAll();
		panSelectedComponent.updateUI();
		
		renderStatement(); 
	}

	public T getEditedFragment() {
		return this.fragment;
	}
}
 
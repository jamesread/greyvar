package greyvarEditor.ui.windows;

import java.awt.FlowLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;

import javax.swing.JDialog;
import javax.swing.JFrame;
import javax.swing.JLabel;
import javax.swing.JSpinner;
import javax.swing.SpinnerNumberModel;

import greyvarEditor.tools.GridResizer;
import jwrCommonsJava.Util;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowResizeGrid extends JDialog {
	public class ResizeResult {
		int w = 0;
		int h = 0; 
	}
	
	private final WindowEditorGrid windowEditorGrid;
	
	public WindowResizeGrid(WindowEditorGrid windowEditorGrid) {
		this.windowEditorGrid = windowEditorGrid;
		  
		setupComponents();
	}
	
	private ResizeResult rr; 
	
	private final JSpinner spinX = new JSpinner(new SpinnerNumberModel(0,0,999,1));
	private final JSpinner spinY = new JSpinner(new SpinnerNumberModel(0,0,999,1)); 
	 
	private void setupComponents() {
		this.setLayout(new GridBagLayout());  
		
		GridBagConstraints gbc = Util.getNewGbc();
		
		this.add(new JLabel("<html>x/width (&rarr;)</html>"), gbc);
		
		gbc.gridx++;
		this.add(spinX, gbc);
		
		gbc.gridy++;
		gbc.gridx = 0;  
		this.add(new JLabel("<html>y/height (&darr;)</html>"), gbc);
		
		gbc.gridx++;  
		this.add(spinY, gbc);
		
		gbc.gridx = 0;
		gbc.gridy++;
		gbc.gridwidth = gbc.REMAINDER;
		gbc.fill = gbc.HORIZONTAL;
		
		this.add(new JButtonWithAl("Done") {
			
			@Override
			public void click() {
				doResize();
				WindowResizeGrid.this.setVisible(false);
			}


		}, gbc);
		
		
		this.setTitle("Resize grid");
		this.setIconImage(WindowMain.getInstance().getIconImage()); 
		this.setLocationRelativeTo(null); 
		this.pack();
		this.doLayout(); 
	}
	
	private void doResize() {
		int x = ((SpinnerNumberModel)spinX.getModel()).getNumber().intValue();
		int y = ((SpinnerNumberModel)spinY.getModel()).getNumber().intValue();
		 
		GridResizer resizer = new GridResizer(windowEditorGrid);
		
		resizer.resize(x, y);
	} 

	public ResizeResult getResult() {
		return this.rr; 
	}
}  

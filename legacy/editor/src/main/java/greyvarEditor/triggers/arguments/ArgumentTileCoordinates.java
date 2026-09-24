package greyvarEditor.triggers.arguments;

import java.awt.FlowLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;

import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JSpinner;
import javax.swing.JTextField;
import javax.swing.SpinnerNumberModel;

import greyvarEditor.triggers.FragmentArgument;
import jwrCommonsJava.Util;

public class ArgumentTileCoordinates extends FragmentArgument {
	public int x = 0;
	public int y = 0;
	
	private class TileCoordEditor extends JPanel {
		private final JSpinner txtX = new JSpinner(new SpinnerNumberModel(0,0,999,1));
		private final JSpinner txtY = new JSpinner(new SpinnerNumberModel(0,0,999,1));
		 
		public TileCoordEditor() {
			this.setup();
			this.refresh();
		}
		
		public void refresh() {
			txtX.setValue(ArgumentTileCoordinates.this.x);
			txtY.setValue(ArgumentTileCoordinates.this.y);
		} 
		
		public int getXValue() { 
			return Integer.parseInt(txtX.getValue().toString());
		}
		
		public int getYValue() { 
			return Integer.parseInt(txtY.getValue().toString());
		}
		
		public void setup() {
			this.setLayout(new GridBagLayout());
			
			GridBagConstraints gbc = Util.getNewGbc();
			gbc.anchor = gbc.CENTER;  
					  
			this.add(txtX, gbc);
			
			gbc.gridx++;			
			this.add(new JLabel(":"), gbc);
			
			gbc.gridx++; 
			this.add(txtY, gbc);
		}
	}
	
	private TileCoordEditor editor = new TileCoordEditor();
	
	public ArgumentTileCoordinates() {
	}
	
	public ArgumentTileCoordinates(int x, int y) {
		this.x = x;
		this.y = y; 
	}
	
	@Override
	public String toString() { 
		return "Tile: {" + this.x + ", " + this.y + "}";
	}
	
	@Override
	public JComponent getEditor() {
		editor.refresh(); 
		return editor; 
	} 
	
	public void setFromEditor() {
		this.x = this.editor.getXValue();
		this.y = this.editor.getYValue();
	}
} 

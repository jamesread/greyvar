package greyvarEditor.triggers.arguments;

import javax.swing.JComponent;
import javax.swing.JTextField;

import greyvarEditor.triggers.FragmentArgument;
 
public class ArgumentString extends FragmentArgument {
	public String value = "hello";
	
	private JTextField txtEditor = new JTextField();

	@Override
	public void setFromEditor() {
		value = txtEditor.getText();   
	}
	
	@Override
	public String toString() {
		return "{" + value + "}";
	} 
	
	@Override
	public JComponent getEditor() {
		txtEditor.setText(value);
		return txtEditor;
	}
} 

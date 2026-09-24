package greyvarEditor.ui.windows;

import java.awt.Color;
import java.awt.FlowLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.io.File;

import javax.swing.JButton;
import javax.swing.JInternalFrame;
import javax.swing.JLabel;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;

import greyvarEditor.dataModel.World;
import greyvarEditor.files.GridFile;
import greyvarEditor.files.GridFileLoader;
import greyvarEditor.files.WorldFileHandler;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowEditorWorld extends JInternalFrame  {
	private World world;

	public WindowEditorWorld(File file) {
		this(WorldFileHandler.load(file));  
	} 
	
	public WindowEditorWorld(final World worldFile) {
		this.world = worldFile;  
		 
		this.setupMenu();
		
		this.setTitle("World Editor:" + worldFile.title); 
		this.setBounds(100, 100, 240, 100); 
		this.setResizable(true);
		this.setMaximizable(true); 
		this.setClosable(true);
		
		JPanel panStuff = new JPanel();
		
		panStuff.setLayout(new GridBagLayout());
		  
		GridBagConstraints gbc = new GridBagConstraints(0,0,1,1,1,1, GridBagConstraints.CENTER, GridBagConstraints.BOTH, new Insets(6,6,6,6), 0, 0); 
		gbc.weighty = 0;
		gbc.anchor = gbc.NORTH;
		gbc.fill = gbc.HORIZONTAL; 
		 
		panStuff.add(new JLabel("Metadata"), gbc); 
		 
		gbc.gridy++; 
		panStuff.add(new JLabel("Author: " + worldFile.author), gbc);
		 
		gbc.gridx++;
		panStuff.add(new JLabel("Spawn: " + worldFile.spawnGrid), gbc);
		
		gbc.gridwidth = gbc.REMAINDER;
		gbc.gridx = 0; 
		gbc.gridy++;  
		panStuff.add(new JButtonWithAl("Triggers") {
			@Override
			public void click() { 
				new WindowTriggerList(WindowMain.getInstance(), world).setVisible(true);
			}
		}, gbc);
		 
		gbc.gridy++; 
		panStuff.add(new JLabel("Grids"), gbc);
		
		for (final File grid : worldFile.gridFiles) {
			if (!grid.getName().endsWith(".grid")) {
				continue;
			} 
			
			JButton btnGrid = new JButton(grid.getName());
			btnGrid.setHorizontalAlignment(SwingConstants.LEFT); 
			btnGrid.addActionListener(new ActionListener() { 
				@Override
				public void actionPerformed(ActionEvent e) { 
					GridFile gf = GridFileLoader.load(grid); 	 	 		 
					
					WindowMain.getInstance().addFrame(new WindowEditorGrid(gf));
				}
			});
			
			gbc.gridy++; 
			gbc.ipadx = 60;
			panStuff.add(btnGrid, gbc); 
		}
		 
		gbc.ipadx = 0; 
		gbc.gridy++;   
		gbc.anchor = gbc.NORTH;  
		gbc.fill = gbc.HORIZONTAL; 
		gbc.weighty = 1; 
		panStuff.add(new JLabel("end"), gbc);
		
		panStuff.setOpaque(true);
		panStuff.setBackground(Color.WHITE);
		
		JScrollPane jsp = new JScrollPane(panStuff);
		jsp.setVerticalScrollBarPolicy(JScrollPane.VERTICAL_SCROLLBAR_ALWAYS); 
		 
		this.add(jsp);  
		
		this.pack();
		this.setVisible(true);
	}
	
	private void setupMenu() {
		JMenuItem mniSave = new JMenuItem("Save");
		mniSave.addActionListener((e) -> {
			onSave();
		}); 
		
		JMenu mnu = new JMenu();  
		mnu.add(mniSave);
	}
	 
	private void onSave() {
		WorldFileHandler.save(this.world);
	}

}

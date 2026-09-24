package greyvarEditor.ui.windows;

import greyvarEditor.GameResources;

import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.Util;
import jwrCommonsJava.ui.ComponentDirectoryChooser;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowOptions extends PopableWindow {
	public WindowOptions() {
		this.setTitle("Options");
		this.setBounds(100, 100, 320, 240);

		this.setupComponents();

		this.setVisible(true);
	}

	private void setupComponents() {
		this.setLayout(new GridBagLayout());
		GridBagConstraints gbc = Util.getNewGbc();
		gbc.weighty = 0;
		gbc.anchor = GridBagConstraints.NORTHWEST;

		this.add(new ComponentDirectoryChooser("Texture directory") {
			@Override
			protected String getInitialValue() {
				return Configuration.getS("paths.textureDirectory");
			}

			@Override
			public void onSelectionChanged() {
				Configuration.set("paths.textureDirectory", this.getSelectedFile().getAbsolutePath());
				GameResources.dirTextures = this.getSelectedFile();
			}
		}, gbc);

		gbc.gridy++;
		this.add(new ComponentDirectoryChooser("Server dat directory") {
			@Override
			protected String getInitialValue() {
				return Configuration.getS("paths.serverDatDirectory");
			}

			@Override
			public void onSelectionChanged() {
				Configuration.set("paths.serverDatDirectory", this.getSelectedFile().getAbsolutePath());
			}
		}, gbc);

		gbc.weighty = 1;
		gbc.gridy++;
		gbc.fill = 0;
		gbc.anchor = GridBagConstraints.SOUTHEAST;
		this.add(new JButtonWithAl("Close") {

			@Override
			public void click() {
				Configuration.save();
				WindowOptions.this.setVisible(false);
			}
		}, gbc);
	}
}
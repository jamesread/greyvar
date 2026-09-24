package greyvarEditor.ui.windows.editors.grid.panels;

import greyvarEditor.Tile;
import greyvarEditor.Tile.TeleportDirection;
import greyvarEditor.files.GridFileLoader;
import greyvarEditor.ui.components.AllEnabledOrDisabledPanel;
import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.WindowMain;

import java.awt.Component;
import java.awt.FlowLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.swing.BorderFactory;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JSpinner;
import javax.swing.JTextField;
import javax.swing.SpinnerNumberModel;

import jwrCommonsJava.Logger;
import jwrCommonsJava.ui.JButtonWithAl;
 
public class PanTransport extends AllEnabledOrDisabledPanel {
	private class ButtonTransportDirection extends JButtonWithAl {
		private int x = 0, y = 0;
		private TeleportDirection direction = TeleportDirection.NORTH;

		public ButtonTransportDirection(int x, int y) {
			super("?");

			if ((x == 0) && (y == 1)) {
				this.setText("<html>&uarr;</html>");
				this.direction = TeleportDirection.NORTH;
			} else if ((x == 1) && (y == 0)) {
				this.setText("<html>&rarr;</html>");
				this.direction = TeleportDirection.EAST;
			} else if ((x == 0) && (y == -1)) {
				this.setText("<html>&darr;</html>");
				this.direction = TeleportDirection.SOUTH;
			} else if ((x == -1) && (y == 0)) {
				this.setText("<html>&larr;</html>");
				this.direction = TeleportDirection.WEST;
			}

			this.x = x;
			this.y = y;
		}
 
		@Override
		public void click() {
			PanTransport.this.cboDirection.setSelectedItem(this.direction);

			int x = PanTransport.this.selectedTileX + this.x;
			x = (x >= 16) ? 0 : x; // FIXME 16 -> Grid Size 
			PanTransport.this.spnDestinationX.setValue(x);

			int y = PanTransport.this.selectedTileY + this.y;
			y = (y >= 16) ? 0 : y; // FIXME 16 -> Grid Size
			PanTransport.this.spnDestinationY.setValue(y);
		}
	}

	private final JTextField txtMessage = new JTextField();
	private final JComboBox cboDirection = new JComboBox(TeleportDirection.values());
	private final JCheckBox chkTeleport = new JCheckBox();
	private final JTextField txtDestinationGrid = new JTextField();
	private final JSpinner spnDestinationX = new JSpinner(new SpinnerNumberModel(0, 0, 16, 1));
	private final JButton btnOpenDestinationGrid = new JButtonWithAl("Open") {
 
		@Override
		public void click() {
			if (!PanTransport.this.txtDestinationGrid.getText().isEmpty()) { 
				WindowMain.getInstance().addFrame(new WindowEditorGrid(GridFileLoader.load(PanTransport.this.txtDestinationGrid.getText())));
			}
		}
	};

	private final JSpinner spnDestinationY = new JSpinner(new SpinnerNumberModel(0, 0, 16, 1));

	private Tile selectedTile;

	private final JButton btnApply = new JButtonWithAl("Apply") {

		@Override
		public void click() {
			PanTransport.this.selectedTile.teleportDestinationX = Integer.parseInt(PanTransport.this.spnDestinationX.getValue().toString());
			PanTransport.this.selectedTile.teleportDestinationY = Integer.parseInt(PanTransport.this.spnDestinationY.getValue().toString());
			PanTransport.this.selectedTile.teleportDestinationGrid = PanTransport.this.txtDestinationGrid.getText();
			PanTransport.this.selectedTile.teleportDirection = TeleportDirection.fromString(PanTransport.this.cboDirection.getSelectedItem().toString());
			PanTransport.this.selectedTile.message = PanTransport.this.txtMessage.getText();
		}
	};

	private int selectedTileX, selectedTileY;

	JPanel panHelperButtons = new JPanel(new FlowLayout());

	public PanTransport() {
		this.setBorder(BorderFactory.createTitledBorder("Transport"));

		this.setupComponents();
	}

	private void setupComponents() {
		this.setLayout(new GridBagLayout());
		GridBagConstraints gbc = new GridBagConstraints(0, 0, 1, 1, 0, 0, GridBagConstraints.NORTHWEST, 1, new Insets(6, 6, 6, 6), 0, 0);

		gbc.anchor = GridBagConstraints.NORTHWEST;
		gbc.weightx = 0;
		gbc.gridx = 0;
		gbc.gridy = 0;
		gbc.gridwidth = 1;
		this.add(new JLabel("Teleport?"), gbc);
		gbc.gridx++;
		gbc.weightx = 1;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.chkTeleport.addActionListener(new ActionListener() {
			@Override
			public void actionPerformed(ActionEvent e) {
				PanTransport.this.teleportCheckboxChanged();
			}
		});
		this.add(this.chkTeleport, gbc);
		this.chkTeleport.setSelected(false);

		gbc.weightx = 0;
		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridwidth = 1;
		this.add(new JLabel("Destination grid"), gbc);
		gbc.gridx++;
		gbc.weightx = 1;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.txtDestinationGrid.setEnabled(false);
		this.add(this.txtDestinationGrid, gbc);

		gbc.gridy++;
		this.add(this.btnOpenDestinationGrid, gbc);

		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridwidth = 1;

		this.add(new JLabel("(buttons:)"), gbc);
		gbc.gridx++;
		this.add(this.panHelperButtons, gbc);
		this.panHelperButtons.add(new ButtonTransportDirection(1, 0));
		this.panHelperButtons.add(new ButtonTransportDirection(0, 1));
		this.panHelperButtons.add(new ButtonTransportDirection(-1, 0));
		this.panHelperButtons.add(new ButtonTransportDirection(0, -1));

		gbc.weightx = 0;
		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridwidth = 1;
		this.add(new JLabel("Direction: "), gbc);
		gbc.gridx++;
		gbc.weightx = 1;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.add(this.cboDirection, gbc);

		gbc.weightx = 0;
		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridwidth = 1;
		this.add(new JLabel("Destination X"), gbc);
		gbc.gridx++;
		gbc.weightx = 1;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.add(this.spnDestinationX, gbc);

		gbc.weightx = 0;
		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridwidth = 1;
		this.add(new JLabel("Destination Y"), gbc);
		gbc.gridx++;
		gbc.weightx = 1;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.add(this.spnDestinationY, gbc);

		gbc.gridx = 0;
		gbc.gridy++;
		this.add(new JLabel("Message"), gbc);

		gbc.gridx++;
		gbc.gridwidth = GridBagConstraints.REMAINDER;
		this.add(this.txtMessage, gbc);

		gbc.gridy++;
		gbc.gridx = 0;
		gbc.gridy++;
		this.add(this.btnApply, gbc);

		this.chkTeleport.getActionListeners()[0].actionPerformed(null);
	}

	private void teleportCheckboxChanged() {
		Logger.messageDebug("cbo teleport: " + PanTransport.this.chkTeleport.isSelected());

		boolean isEnabled = PanTransport.this.chkTeleport.isSelected();

		this.txtDestinationGrid.setEnabled(isEnabled);
		this.btnOpenDestinationGrid.setEnabled(isEnabled);
		this.spnDestinationX.setEnabled(isEnabled);
		this.spnDestinationY.setEnabled(isEnabled);
		this.cboDirection.setEnabled(isEnabled);

		for (Component c : PanTransport.this.panHelperButtons.getComponents()) {
			c.setEnabled(isEnabled);
		}
	}

	public void updateSelectedTile(int x, int y, Tile t) {
		this.selectedTileX = x;
		this.selectedTileY = y;
		this.selectedTile = t;

		this.chkTeleport.setSelected(!this.selectedTile.teleportDestinationGrid.isEmpty());
		this.txtDestinationGrid.setText(t.teleportDestinationGrid);
		this.spnDestinationX.setValue(t.teleportDestinationX);
		this.spnDestinationY.setValue(t.teleportDestinationY);
		this.txtMessage.setText(t.message);

		Logger.messageDebug("SI" + t.teleportDirection);
		this.cboDirection.setSelectedItem(t.teleportDirection);

		this.teleportCheckboxChanged();
	}
}

(function() {
	// Function to get query parameters
	// Note: This app uses ? as separator between params (not standard &)
	var urlParam = function(name, w) {
		w = w || window;
		var rx = new RegExp('[\&|\?]' + name + '=([^\&\#\?]+)'),
			val = w.location.search.match(rx);
		return !val ? '' : val[1];
	};

	// Retrieve the RouterID and RouterName query parameters
	var routerId = urlParam('RouterID');
	var routerNameForTitle = urlParam('RouterName');

	console.log("routerId:", routerId);
	console.log("routerNameForTitle:", routerNameForTitle);

	document.title = `TopoViewer::${routerNameForTitle || routerId}`;


	// Initialize the terminal with the desired options
	var terminal = new Terminal({
		screenKeys: true,
		useStyle: true,
		cursorBlink: true,
		fullscreenWin: true,
		maximizeWin: true,
		screenReaderMode: true,
		cols: 128,
	});

	// Open the terminal in the specified HTML element
	terminal.open(document.getElementById("terminal"));

	// Determine the WebSocket protocol based on the page's protocol
	var protocol = (location.protocol === "https:") ? "wss://" : "ws://";
	var url = protocol + location.host + "/xterm.js";
	var ws = new WebSocket(url);

	console.log(ws);

	// Load necessary addons
	var attachAddon = new AttachAddon.AttachAddon(ws);
	var fitAddon = new FitAddon.FitAddon();
	var webLinksAddon = new WebLinksAddon.WebLinksAddon();
	var unicode11Addon = new Unicode11Addon.Unicode11Addon();
	var serializeAddon = new SerializeAddon.SerializeAddon();

	terminal.loadAddon(fitAddon);
	terminal.loadAddon(webLinksAddon);
	terminal.loadAddon(unicode11Addon);
	terminal.loadAddon(serializeAddon);

	// Define WebSocket event handlers
	ws.onclose = function(event) {
		console.log(event);
		terminal.write('\r\n\nconnection has been terminated from the server-side (hit refresh to restart)\n');
	};

	ws.onopen = function() {
		terminal.loadAddon(attachAddon);
		terminal._initialized = true;
		terminal.focus();
		setTimeout(function() {
			fitAddon.fit();
		});

		// Get parameters from URL (URL uses ? as separator, so we need to clean up)
		var routerNameRaw = urlParam('RouterName');
		var routerName = routerNameRaw.split("?")[0];  // Remove trailing ?Kind=... if present
		var nodeKind = urlParam('Kind') || 'default';
		var nodeImage = decodeURIComponent(urlParam('Image') || '');

		// Build command based on node kind
		// Reference: https://containerlab.dev/manual/kinds/
		var terminalCommand;
		if (nodeKind === 'linux' && nodeImage.toLowerCase().includes('frr')) {
			// FRRouting: use docker exec with vtysh for direct FRR CLI access
			terminalCommand = "docker exec -it " + routerName + " vtysh";
		} else if (nodeKind === 'linux') {
			// Linux containers: use docker exec for direct access
			terminalCommand = "docker exec -it " + routerName + " bash";
		} else if (nodeKind === 'cisco_iol') {
			// Cisco IOL: admin/admin
			terminalCommand = "sshpass -p 'admin' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else if (nodeKind === 'arista_ceos') {
			// Arista cEOS: admin/admin
			terminalCommand = "sshpass -p 'admin' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else if (nodeKind === 'juniper_crpd') {
			// Juniper cRPD: root/clab123
			terminalCommand = "sshpass -p 'clab123' ssh -q -o StrictHostKeyChecking=no root@" + routerName;
		} else if (nodeKind === 'paloalto_panos') {
			// Palo Alto PAN: admin/Admin@123
			terminalCommand = "sshpass -p 'Admin@123' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else if (nodeKind === 'nokia_srlinux') {
			// Nokia SR Linux: admin/NokiaSrl1!
			terminalCommand = "sshpass -p 'NokiaSrl1!' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else if (nodeKind === 'vyosnetworks_vyos') {
			// VyOS: admin/admin
			terminalCommand = "sshpass -p 'admin' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else if (nodeKind === 'fortinet_fortigate') {
			// Fortinet FortiGate: admin/Admin123!@#$
			terminalCommand = "sshpass -p 'Admin123!@#$' ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		} else {
			// Other network devices: use SSH with admin user (default fallback, no auto-password)
			terminalCommand = "ssh -q -o StrictHostKeyChecking=no admin@" + routerName;
		}
		console.log("Terminal command: " + terminalCommand + " (kind: " + nodeKind + ", image: " + nodeImage + ")");
		ws.send(terminalCommand + "\n");



		terminal.onResize(function(event) {
			var rows = event.rows;
			var cols = event.cols;
			var size = JSON.stringify({
				cols: cols,
				rows: rows + 1
			});
			var send = new TextEncoder().encode("\x01" + size);
			console.log('resizing to', size);


			ws.send(send);
			fitAddon.fit(); // this code indeedd
		});

		terminal.onTitleChange(function(event) {
			console.log(event);
		});

		// Fit the terminal to the window size when the window is resized
		window.onresize = function() {
			fitAddon.fit();
		};
	};
})();
use tauri_plugin_shell::ShellExt;
use tauri_plugin_shell::process::CommandEvent;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let sidecar_command = app.shell().sidecar("tablepilot").unwrap();
            let (mut rx, mut _child) = sidecar_command
                .args(["serve"])
                .spawn()
                .expect("Failed to start tablepilot server");

	    tauri::async_runtime::spawn(async move {
		// read events such as stdout
		while let Some(event) = rx.recv().await {
		    if let CommandEvent::Stderr(line_bytes) = event {
			let line = String::from_utf8_lossy(&line_bytes);
			println!("Tablepilot server: {}", line);
		    }
		}
	    });
            if cfg!(debug_assertions) {
                app.handle().plugin(
                    tauri_plugin_log::Builder::default()
                        .level(log::LevelFilter::Info)
                        .build(),
                )?;
            }
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

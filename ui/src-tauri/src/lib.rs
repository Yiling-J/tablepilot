use std::sync::{Arc, Mutex};

use std::fs;
use tauri::Manager;
use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(
            tauri_plugin_log::Builder::new()
                .target(tauri_plugin_log::Target::new(
                    tauri_plugin_log::TargetKind::LogDir {
                        file_name: Some("logs".to_string()),
                    },
                ))
                .build(),
        )
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let dir = app
                .path()
                .resolve("", tauri::path::BaseDirectory::AppData)
                .unwrap();
            let config_file_path = app
                .path()
                .resolve("config.toml", tauri::path::BaseDirectory::Resource)
                .unwrap();

            let dest_path = dir.join("config.toml");
            if let Some(parent) = dest_path.parent() {
                fs::create_dir_all(parent).expect("failed to create directories");
            }
            fs::copy(&config_file_path, &dest_path).expect("failed to copy config.toml");

            let sidecar_command = app.shell().sidecar("tablepilot").unwrap();
            let (mut rx, child) = sidecar_command
                .args(["serve"])
                .current_dir(dir)
                .spawn()
                .unwrap();

            tauri::async_runtime::spawn(async move {
                // read events such as stdout
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line_bytes) => {
                            let line = String::from_utf8_lossy(&line_bytes);
                            log::info!("Tablepilot server (stdout): {}", line);
                        }
                        CommandEvent::Stderr(line_bytes) => {
                            let line = String::from_utf8_lossy(&line_bytes);
                            log::error!("Tablepilot server (stderr): {}", line);
                        }
                        _ => {} // Handle any other events if needed
                    }
                }
            });

            let window = app.get_webview_window("main").unwrap();
            let arc_child = Arc::new(Mutex::new(Some(child)));
            let child_clone = Arc::clone(&arc_child);

            window.on_window_event(move |event| {
                if let tauri::WindowEvent::CloseRequested { .. } = event {
                    let mut child_lock = child_clone.lock().unwrap();
                    if let Some(mut child_process) = child_lock.take() {
                        if let Err(e) = child_process.write("Exit message from Rust\n".as_bytes()) {
                            println!("Fail to send to stdin of Tablepilot: {}", e);
                        }

                        if let Err(e) = child_process.kill() {
                            eprintln!("Failed to kill child process: {}", e);
                        }
                    }
                }
            });
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

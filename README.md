# osdeps
OS Dependency Scanner

# Problem
Certain languages, e.g. Python, provide ways to pull in dependencies which may link to libraries in the operating system's environment.

`osdeps-cli` exists to enumerate the operating system dependencies required for packages installed within a Python environment.

# Motivation
I wanted to build a tool to help with deploying software to cloud environments. I noticed a pain point in cloud-based software development cycles in that debugging deployment issues usually takes a long time. What I'm referring to is having to repeatedly deploy an app to solve some problem which can only be debugged from the environment in which it is deployed. I think this tool can help these situations by ultimately cutting down on one reason for these types of cycles -- **figuring out what the system level dependencies for your app should be in the environment in which it is deployed**.

 The tool leverages a static analysis approach to check system level dependencies for Python applications. What this gets you is a way to know if your Python application will have errors related to missing or incorrectly versioned operating system libraries before deploying your app. Ideally this manifest would be used to correct any issues during deployment of the container after `pip install...`.

# Moving Forward
This tool could be expanded to support Windows and JavaScript (Node.js).

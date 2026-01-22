# Quickstart: Package-Mode Agent Installation

**Feature**: 001-package-mode-support
**Date**: 2026-01-22
**Audience**: System administrators deploying Flight Control on package-managed RHEL/Ubuntu systems

## Overview

This guide covers installing and configuring the Flight Control agent on traditional Linux distributions where the OS is managed via package managers (dnf/apt) rather than bootc image updates. The agent will manage Flight Control configurations and applications while leaving OS updates to your existing package management workflows.

## Prerequisites

### System Requirements

**RHEL**:
- RHEL 9.0 or later
- systemd
- Active subscription or access to RHEL repositories
- Podman (for container workload management)

**Ubuntu**:
- Ubuntu 22.04 LTS or later
- systemd
- Podman (install via `apt install podman`)

### Network Requirements

- Outbound HTTPS (443) to Flight Control management service
- DNS resolution for management service hostname
- (Optional) Outbound HTTPS to container registries (if managing containerized applications)

### Not Required

- bootc installation (agent detects absence and operates in package-mode)
- OS image build infrastructure
- Image registry credentials for OS images

## Installation

### RHEL 9+ Installation

1. **Add Flight Control Repository** (if not already configured):
   ```bash
   # Example repository configuration (adjust URL for your deployment)
   sudo dnf config-manager --add-repo https://flightctl.example.com/repo/rhel9/flightctl.repo
   ```

2. **Install Agent Package**:
   ```bash
   sudo dnf install flightctl-agent
   ```

3. **Verify Installation**:
   ```bash
   rpm -qi flightctl-agent
   systemctl status flightctl-agent
   ```

### Ubuntu 22.04+ Installation

1. **Add Flight Control Repository** (if not already configured):
   ```bash
   # Example repository configuration (adjust for your deployment)
   curl -fsSL https://flightctl.example.com/repo/ubuntu/flightctl.gpg | \
       sudo gpg --dearmor -o /usr/share/keyrings/flightctl-archive-keyring.gpg

   echo "deb [signed-by=/usr/share/keyrings/flightctl-archive-keyring.gpg] \
       https://flightctl.example.com/repo/ubuntu jammy main" | \
       sudo tee /etc/apt/sources.list.d/flightctl.list
   ```

2. **Update Package Index**:
   ```bash
   sudo apt update
   ```

3. **Install Agent Package**:
   ```bash
   sudo apt install flightctl-agent
   ```

4. **Verify Installation**:
   ```bash
   dpkg -l | grep flightctl-agent
   systemctl status flightctl-agent
   ```

## Configuration

### Initial Configuration

1. **Create Agent Configuration File**:
   ```bash
   sudo mkdir -p /etc/flightctl
   sudo nano /etc/flightctl/agent.yaml
   ```

2. **Minimal Configuration** (`/etc/flightctl/agent.yaml`):
   ```yaml
   server:
     url: https://flightctl-api.example.com
     insecureSkipVerify: false  # Set to true only for testing

   enrollment:
     token: YOUR_ENROLLMENT_TOKEN
     # OR use certificate-based enrollment:
     # certificatePath: /etc/flightctl/certs/agent-cert.pem
     # keyPath: /etc/flightctl/certs/agent-key.pem

   dataDir: /var/lib/flightctl

   logLevel: info
   ```

3. **Set Permissions**:
   ```bash
   sudo chmod 600 /etc/flightctl/agent.yaml
   sudo chown root:root /etc/flightctl/agent.yaml
   ```

### Enrollment

**Option 1: Enrollment Token** (Recommended for initial deployment):

1. Generate token in Flight Control console or via API:
   ```bash
   flightctl enrollment-request create \
       --name rhel9-edge-001 \
       --labels deployment=edge,location=warehouse-5
   ```

2. Copy token to agent config (`/etc/flightctl/agent.yaml`):
   ```yaml
   enrollment:
     token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

**Option 2: Certificate-Based** (For production deployments):

1. Generate CSR:
   ```bash
   openssl req -new -newkey rsa:4096 -nodes \
       -keyout /etc/flightctl/certs/agent-key.pem \
       -out /etc/flightctl/certs/agent-csr.pem \
       -subj "/CN=rhel9-edge-001"
   ```

2. Submit CSR to Flight Control CA (via API or console)

3. Place signed certificate in `/etc/flightctl/certs/agent-cert.pem`

4. Update config to use certificate

### Start Agent

```bash
# Enable agent to start on boot
sudo systemctl enable flightctl-agent

# Start agent
sudo systemctl start flightctl-agent

# Verify agent is running
sudo systemctl status flightctl-agent

# Check agent logs
sudo journalctl -u flightctl-agent -f
```

### Verify Package-Mode Detection

Check agent logs for package-mode confirmation:

```bash
sudo journalctl -u flightctl-agent | grep -i "package-mode"
```

Expected output:
```
Jan 22 10:15:23 rhel9-edge-001 flightctl-agent[1234]: Package-mode detected: bootc not found
```

In Flight Control console, device status should show:
- **Deployment Mode**: Package Mode
- **OS Image**: (empty)

## Configuration Updates

### Applying Configuration via Flight Control

1. **Create Device Configuration** (in console or via API):
   ```yaml
   apiVersion: v1beta1
   kind: Device
   metadata:
     name: rhel9-edge-001
   spec:
     config:
       - path: /etc/myapp/config.yaml
         content: |
           setting1: value1
           setting2: value2
       - path: /etc/motd
         content: |
           Welcome to Flight Control managed device
   ```

2. **Agent Applies Configuration**:
   - Agent polls for configuration changes (default: every 60s)
   - Downloads configuration
   - Writes files to specified paths
   - Logs: "Configuration version 1234 applied"

3. **Verify Application**:
   ```bash
   cat /etc/myapp/config.yaml
   sudo journalctl -u flightctl-agent | grep -i "configuration.*applied"
   ```

## OS Updates (Manual)

**Important**: In package-mode, Flight Control does NOT manage OS updates. Continue using your existing OS update procedures.

### RHEL OS Updates

```bash
# Check for updates
sudo dnf check-update

# Apply OS updates
sudo dnf update -y

# Reboot if kernel updated
sudo reboot
```

### Ubuntu OS Updates

```bash
# Check for updates
sudo apt update
sudo apt list --upgradable

# Apply OS updates
sudo apt upgrade -y

# Reboot if kernel updated
sudo reboot
```

**Flight Control Behavior**: Agent detects reboot (via boot ID change), continues managing configurations and applications after reboot. No conflicts with OS package manager operations.

## Application Management

### Deploy Containerized Application

1. **Create Application Spec** (in console or via API):
   ```yaml
   apiVersion: v1beta1
   kind: Device
   metadata:
     name: rhel9-edge-001
   spec:
     applications:
       - name: nginx-app
         image: docker.io/library/nginx:latest
         envVars:
           - name: PORT
             value: "8080"
   ```

2. **Agent Deploys Application**:
   - Pulls container image via Podman
   - Creates systemd service for application
   - Starts and enables service
   - Reports application status

3. **Verify Deployment**:
   ```bash
   sudo podman ps
   sudo systemctl status flightctl-app-nginx-app
   ```

## Troubleshooting

### Agent Not Starting

**Check systemd status**:
```bash
sudo systemctl status flightctl-agent
```

**Common issues**:
- Configuration file syntax errors: `sudo flightctl-agent validate-config`
- Permission errors: Ensure `/etc/flightctl/agent.yaml` is owned by root with mode 600
- Network connectivity: Test connection to management service: `curl -I https://flightctl-api.example.com`

### Package-Mode Not Detected

**Check bootc presence**:
```bash
which bootc
```

If bootc is installed but you want package-mode:
```bash
# Uninstall bootc (RHEL)
sudo dnf remove bootc

# Uninstall bootc (Ubuntu)
sudo apt remove bootc

# Restart agent
sudo systemctl restart flightctl-agent
```

### OS Update Warnings

If agent logs warnings about OS updates:
```
Ignoring OS update to quay.io/flightctl/os-image: package-mode device
```

This is **expected behavior**. Flight Control spec may include OS image updates for image-mode devices in the same fleet. Package-mode devices safely ignore these updates.

**Action**: No action required. If you want to enable OS image management, install bootc and restart the agent.

### Configuration Not Applying

**Check agent connectivity**:
```bash
sudo journalctl -u flightctl-agent | grep -i "connection\|error"
```

**Verify device enrollment**:
```bash
# Via Flight Control API
curl -H "Authorization: Bearer $TOKEN" \
     https://flightctl-api.example.com/api/v1/devices/rhel9-edge-001
```

**Force configuration sync**:
```bash
# Restart agent to trigger immediate sync
sudo systemctl restart flightctl-agent
```

## Upgrading Agent

### RHEL Agent Upgrade

```bash
# Check for agent updates
sudo dnf check-update flightctl-agent

# Upgrade agent
sudo dnf update flightctl-agent

# Agent auto-updates may also be pushed via Flight Control (if configured)
```

### Ubuntu Agent Upgrade

```bash
# Check for agent updates
sudo apt update
sudo apt list --upgradable | grep flightctl-agent

# Upgrade agent
sudo apt upgrade flightctl-agent
```

## Uninstallation

### Remove Agent (RHEL)

```bash
# Stop and disable agent
sudo systemctl stop flightctl-agent
sudo systemctl disable flightctl-agent

# Remove package
sudo dnf remove flightctl-agent

# (Optional) Remove configuration and data
sudo rm -rf /etc/flightctl /var/lib/flightctl
```

### Remove Agent (Ubuntu)

```bash
# Stop and disable agent
sudo systemctl stop flightctl-agent
sudo systemctl disable flightctl-agent

# Remove package
sudo apt remove flightctl-agent

# (Optional) Purge configuration
sudo apt purge flightctl-agent
sudo rm -rf /var/lib/flightctl
```

## Next Steps

- [Managing Fleets](../../../docs/user/using/managing-fleets.md)
- [Device Observability](../../../docs/user/using/device-observability.md)
- [Troubleshooting Guide](../../../docs/user/using/troubleshooting.md)
- [API Reference](../../../docs/user/references/api-resources.md)

## Differences from Image-Mode

| Aspect | Package-Mode | Image-Mode |
|--------|--------------|------------|
| **OS Updates** | Manual (dnf/apt) | Automated via Flight Control |
| **Agent Updates** | Via package manager OR Flight Control | Via Flight Control |
| **Bootc Requirement** | Not required | Required |
| **Reboot on OS Update** | Manual | Automatic after OS image switch |
| **Configuration Management** | Identical | Identical |
| **Application Management** | Identical | Identical |

Package-mode provides a migration path for existing deployments without requiring adoption of image-based OS management.

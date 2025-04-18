# List targets
iscsiadm  -m discovery -t sendtargets -p 100.100.0.147:3260

# Login
iscsiadm  -m node  -T iqn.2025-04.net.virer.virium:0a062498-f339-45be-82b5-0cca55f19f12 -p 100.100.0.147:3260 -l

# Logout
iscsiadm  -m node  -p 100.100.0.147:3260 -u
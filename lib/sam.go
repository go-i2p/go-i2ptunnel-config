package i2pconv

import (
    "github.com/go-i2p/i2pkeys"
    "path/filepath"
)

// SAMTunnel returns the I2P keys and SAM options for this tunnel configuration.
// If PersistentKeys is true, keys will be stored in a SAMv3 compatible format.
func (c *TunnelConfig) SAMTunnel() (*i2pkeys.I2PKeys, []string, error) {
    var keys *i2pkeys.I2PKeys
    var err error
    
    if c.PersistentKeys {
        // Get default I2P keystore path
        keystore, err := i2pkeys.I2PKeystorePath()
        if err != nil {
            return nil, nil, err
        }
        
        // Load or create keys for this tunnel
        keypath := filepath.Join(keystore, c.Name+".keys")
        loadedKeys, err := i2pkeys.LoadKeys(keypath) 
        if err == nil {
            keys = &loadedKeys
        } else {
            // Create new keys if none exist
            newKeys, err := i2pkeys.GenerateKeys()
            if err != nil {
                return nil, nil, err
            }
            keys = &newKeys
            // Store the new keys
            if err := i2pkeys.StoreKeysIncompat(newKeys, keypath); err != nil {
                return nil, nil, err
            }
        }
    }

    // Generate SAM options from config
    var opts []string
    
    // Process I2CP options
    for k, v := range c.I2CP {
        opts = append(opts, "i2cp."+k+"="+fmt.Sprint(v))
    }

    // Process tunnel options 
    for k, v := range c.Tunnel {
        opts = append(opts, k+"="+fmt.Sprint(v))
    }

    // Process inbound/outbound options
    for k, v := range c.Inbound {
        opts = append(opts, "inbound."+k+"="+fmt.Sprint(v))
    }
    for k, v := range c.Outbound {
        opts = append(opts, "outbound."+k+"="+fmt.Sprint(v))
    }

    // Ensure lease set encryption
    if !hasOption(opts, "i2cp.leaseSetEncType") {
        opts = append(opts, "i2cp.leaseSetEncType=4,0")
    }

    return keys, opts, nil
}

func hasOption(opts []string, prefix string) bool {
    for _, opt := range opts {
        if strings.HasPrefix(opt, prefix) {
            return true
        }
    }
    return false
}

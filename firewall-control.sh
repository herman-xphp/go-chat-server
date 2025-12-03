#!/bin/bash

# Script untuk control firewall port 8080

case "$1" in
    open)
        echo "🔓 Membuka port 8080 (sementara - hilang setelah reboot)"
        sudo firewall-cmd --add-port=8080/tcp
        echo "✅ Port 8080 terbuka untuk test"
        echo "📱 Akses dari HP: http://$(ip addr show | grep "inet " | grep -v "127.0.0.1" | head -1 | awk '{print $2}' | cut -d'/' -f1):8080"
        ;;
    
    close)
        echo "🔒 Menutup port 8080"
        sudo firewall-cmd --remove-port=8080/tcp
        echo "✅ Port 8080 ditutup"
        ;;
    
    permanent-open)
        echo "⚠️  PERINGATAN: Membuka port 8080 PERMANENT!"
        read -p "Yakin? (y/n): " confirm
        if [ "$confirm" = "y" ]; then
            sudo firewall-cmd --add-port=8080/tcp --permanent
            sudo firewall-cmd --reload
            echo "✅ Port 8080 terbuka permanent"
        else
            echo "❌ Dibatalkan"
        fi
        ;;
    
    permanent-close)
        echo "🔒 Menutup port 8080 permanent"
        sudo firewall-cmd --remove-port=8080/tcp --permanent
        sudo firewall-cmd --reload
        echo "✅ Port 8080 ditutup permanent"
        ;;
    
    status)
        echo "📊 Status Firewall:"
        sudo firewall-cmd --list-ports | grep 8080 && echo "✅ Port 8080 TERBUKA" || echo "🔒 Port 8080 TERTUTUP"
        ;;
    
    *)
        echo "Usage: $0 {open|close|permanent-open|permanent-close|status}"
        echo ""
        echo "Commands:"
        echo "  open              - Buka port sementara (untuk test)"
        echo "  close             - Tutup port sementara"
        echo "  permanent-open    - Buka port permanent (dengan konfirmasi)"
        echo "  permanent-close   - Tutup port permanent"
        echo "  status            - Cek status port"
        echo ""
        echo "Contoh:"
        echo "  $0 open           # Buka untuk test"
        echo "  $0 close          # Tutup setelah test"
        exit 1
        ;;
esac

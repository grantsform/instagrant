# Instagrant development justfile
# NOTE: VM functionality has been moved to the CLI: ./instagrant vm

# Default recipe
default:
    @echo "VM functionality is now in the CLI:"
    @echo "  ./instagrant vm setup  - Setup and boot Arch ISO"
    @echo "  ./instagrant vm boot   - Boot from installed disk"
    @echo "  ./instagrant vm check  - Check disk image"
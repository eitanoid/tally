mod_dir := module_directory()

# Build and install debug APK onto physical device
install:
    set working-directory = {{ mod_dir }}
    cd android && ./gradlew installDebug --max-workers=2

# Start Metro JS bundler with restricted memory
metro:
    set working-directory = {{ mod_dir }}
    NODE_OPTIONS="--max-old-space-size=2048" yarn android

# Clean Android build cache and stop Gradle daemons
clean:
    set working-directory = {{ mod_dir }}
    cd android && ./gradlew clean --stop

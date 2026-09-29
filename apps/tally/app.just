mod_dir := module_directory()

# Start Metro JS bundler with restricted memory
android:
    set working-directory = {{ mod_dir }}
    npx expo run:android

# Clean Android build cache and stop Gradle daemons
clean:
    set working-directory = {{ mod_dir }}
    cd android && ./gradlew clean --stop

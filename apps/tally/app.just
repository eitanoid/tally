mod_dir := module_directory()

install:
    set working-directory = {{ mod_dir }}
    yarn install

build-modules:
    set working-directory = {{ mod_dir }}
    npx expo prebuild --clean --platform android

# Build and package app for android
android:
    set working-directory = {{ mod_dir }}
    npx expo run:android

# Clean Android build cache and stop Gradle daemons
clean:
    set working-directory = {{ mod_dir }}
    cd android && ./gradlew clean --stop

doctor:
    npx expo doctor

build-apk:
    npx eas-cli build --platform android --profile development --local

adb:
    @echo 'to paid a phone run:'
    @echo 'adb pair <ip>:<port>'
    @echo 'adb connect <ip>:<port>'
    @echo 'adb connect <ip>:<port>'
    @echo 'adb reverse tcp:8081 tcp:8081 <ip>'

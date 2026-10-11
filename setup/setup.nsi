!include "MUI2.nsh"

!ifndef VERSION
  !define VERSION "dev"
!endif

!define NAME "Saab NG9-3 Halogen Check Disabler"
!define REGPATH_UNINSTSUBKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"

Name "${NAME}"
OutFile "../hcd_setup.exe"
Unicode True
RequestExecutionLevel Admin

InstallDir "$ProgramFiles\Roffe\HalogenCheckDisabler"
InstallDirRegKey HKLM "${REGPATH_UNINSTSUBKEY}" "UninstallString"

Var StartMenuFolder

!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY

!define MUI_STARTMENUPAGE_REGISTRY_ROOT "HKLM"
!define MUI_STARTMENUPAGE_REGISTRY_KEY "Software\HalogenCheckDisabler"
!define MUI_STARTMENUPAGE_REGISTRY_VALUENAME "Start Menu Folder"

!insertmacro MUI_PAGE_STARTMENU Application $StartMenuFolder
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

!macro EnsureAdminRights
  UserInfo::GetAccountType
  Pop $0
  ${If} $0 != "admin"
    MessageBox MB_IconStop "Administrator rights required!"
    SetErrorLevel 740 ; ERROR_ELEVATION_REQUIRED
    Quit
  ${EndIf}
!macroend

Function .onInit
  SetShellVarContext All
  !insertmacro EnsureAdminRights
FunctionEnd

Function un.onInit
  SetShellVarContext All
  !insertmacro EnsureAdminRights
FunctionEnd

Section "${NAME}" SecCore
  SectionIn RO
  SetOutPath "$InstDir"
  File "../Saab NG9-3 Halogen Check Disabler.exe"

  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "DisplayName" "${NAME}"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "Publisher" "roffe"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "URLInfoAbout" "https://roffe.nu/"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "DisplayIcon" "$InstDir\${NAME}.exe,0"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "UninstallString" '"$InstDir\Uninstall.exe"'
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "QuietUninstallString" '"$InstDir\Uninstall.exe" /S'
  WriteRegDWORD HKLM "${REGPATH_UNINSTSUBKEY}" "NoModify" 1
  WriteRegDWORD HKLM "${REGPATH_UNINSTSUBKEY}" "NoRepair" 1

  WriteUninstaller "$InstDir\Uninstall.exe"

  !insertmacro MUI_STARTMENU_WRITE_BEGIN Application
    CreateDirectory "$SMPROGRAMS\$StartMenuFolder"
    CreateShortcut "$SMPROGRAMS\$StartMenuFolder\Halogen Check Disabler.lnk" "$InstDir\${NAME}.exe"
    CreateShortcut "$SMPROGRAMS\$StartMenuFolder\Uninstall.lnk" "$InstDir\Uninstall.exe"
  !insertmacro MUI_STARTMENU_WRITE_END

  CreateShortcut "$Desktop\Halogen Check Disabler.lnk" "$InstDir\${NAME}.exe"
SectionEnd

Section "Tech2 SAE J2534 DLL" SecTech2
  SetOutPath "$InstDir\Tech2 SAE J2534 DLL"
  File /r "../dist/Tech2 SAE J2534 DLL/"
  ExecWait 'msiexec /i "$InstDir\Tech2 SAE J2534 DLL\Tech2 SAE J2534 DLL.msi" /qn'
SectionEnd

Section "Software OpenGL (Mesa)" SecOpenGL
  SetOutPath "$InstDir"
  File "../dist/opengl32.dll"
SectionEnd

LangString DESC_SecCore ${LANG_ENGLISH} "The main program."
LangString DESC_SecTech2 ${LANG_ENGLISH} "Tech2 SAE J2534 passthru driver (installed silently)."
LangString DESC_SecOpenGL ${LANG_ENGLISH} "Software OpenGL renderer for machines without working GPU drivers."

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} $(DESC_SecCore)
  !insertmacro MUI_DESCRIPTION_TEXT ${SecTech2} $(DESC_SecTech2)
  !insertmacro MUI_DESCRIPTION_TEXT ${SecOpenGL} $(DESC_SecOpenGL)
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Section "Uninstall"
  Delete "$InstDir\${NAME}.exe"
  Delete "$InstDir\opengl32.dll"
  Delete "$InstDir\Uninstall.exe"
  RMDir /r "$InstDir\Tech2 SAE J2534 DLL"
  RMDir "$InstDir"

  !insertmacro MUI_STARTMENU_GETFOLDER Application $StartMenuFolder
  Delete "$SMPROGRAMS\$StartMenuFolder\Halogen Check Disabler.lnk"
  Delete "$SMPROGRAMS\$StartMenuFolder\Uninstall.lnk"
  RMDir "$SMPROGRAMS\$StartMenuFolder"
  Delete "$Desktop\Halogen Check Disabler.lnk"

  DeleteRegKey HKLM "${REGPATH_UNINSTSUBKEY}"
  DeleteRegKey /ifempty HKLM "Software\HalogenCheckDisabler"
SectionEnd

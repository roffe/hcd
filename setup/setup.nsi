;--------------------------------
;Include Modern UI

  !include "MUI2.nsh"

;--------------------------------
;General
  !define NAME "LightControl"
  !define REGPATH_UNINSTSUBKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"

  ;Name and file
  Name "${NAME}"
  OutFile ..\lightcontrol_setup.exe
  Unicode True
  RequestExecutionLevel Admin ; Request admin rights on WinVista+ (when UAC is turned on)
  
  ;Default installation folder
  InstallDir "$ProgramFiles\Mackanized\$(^Name)"
  
  ;Get installation folder from registry if available
  InstallDirRegKey HKLM "${REGPATH_UNINSTSUBKEY}" "UninstallString"

;--------------------------------
;Variables
  Var StartMenuFolder
;--------------------------------
;Interface Settings
  !define MUI_ABORTWARNING
  !define MUI_HEADERIMAGE
  !define MUI_HEADERIMAGE_BITMAP "header.bmp"
;--------------------------------
;Pages
  ; !insertmacro MUI_PAGE_LICENSE "${NSISDIR}\Docs\Modern UI\License.txt"
  !insertmacro MUI_PAGE_COMPONENTS
  !insertmacro MUI_PAGE_DIRECTORY
  
  ;Start Menu Folder Page Configuration
  !define MUI_STARTMENUPAGE_REGISTRY_ROOT "HKLM" 
  !define MUI_STARTMENUPAGE_REGISTRY_KEY "Software\LightControl" 
  !define MUI_STARTMENUPAGE_REGISTRY_VALUENAME "Start Menu Folder"
  
  !insertmacro MUI_PAGE_STARTMENU Application $StartMenuFolder
  
  !insertmacro MUI_PAGE_INSTFILES
  
  !insertmacro MUI_UNPAGE_CONFIRM
  !insertmacro MUI_UNPAGE_INSTFILES

;--------------------------------
;Languages
  !insertmacro MUI_LANGUAGE "English"
;--------------------------------
;Installer Sections

!macro EnsureAdminRights
  UserInfo::GetAccountType
  Pop $0
  ${If} $0 != "admin" ; Require admin rights on WinNT4+
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


Section "LightControl" SecCore
  SetOutPath "$InstDir"

  ;ADD YOUR OWN FILES HERE...
  FILE ..\LightControl.exe
  FILE ..\dist\readme.txt

  ;Store installation folder
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "DisplayName" "${NAME}"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "DisplayIcon" "$InstDir\LightControl.exe,0"
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "UninstallString" '"$InstDir\Uninstall.exe"'
  WriteRegStr HKLM "${REGPATH_UNINSTSUBKEY}" "QuietUninstallString" '"$InstDir\Uninstall.exe" /S'
  WriteRegDWORD HKLM "${REGPATH_UNINSTSUBKEY}" "NoModify" 1
  WriteRegDWORD HKLM "${REGPATH_UNINSTSUBKEY}" "NoRepair" 1
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers" "$InstDir\LightControl.exe" "RUNASADMIN" 
  
  ;Create uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  
  !insertmacro MUI_STARTMENU_WRITE_BEGIN Application
    ;Create shortcuts
    CreateDirectory "$SMPROGRAMS\$StartMenuFolder"
    CreateShortcut "$SMPROGRAMS\$StartMenuFolder\LightControl.lnk" "$INSTDIR\LightControl.exe"
    CreateShortcut "$SMPROGRAMS\$StartMenuFolder\Uninstall.lnk" "$INSTDIR\Uninstall.exe"
  !insertmacro MUI_STARTMENU_WRITE_END
SectionEnd

Section /o "Tech2 SAE J2534 DLL" SecTech2
  SetOutPath "$InstDir"
  FILE /r "..\dist\Tech2 SAE J2534 DLL"

  ;Run bundled setup.exe (Tech2 J2534 driver installer)
  ExecWait '"$InstDir\Tech2 SAE J2534 DLL\setup.exe"'
SectionEnd

Section /o "OpenGL dll" SecOpenGL
  SetOutPath "$InstDir"
  FILE ..\dist\opengl32.dll
SectionEnd

;--------------------------------
;Descriptions

  ;Language strings
  LangString DESC_SecCore ${LANG_ENGLISH} "LightControl Core."
  LangString DESC_SecTech2 ${LANG_ENGLISH} "Tech2 SAE J2534 DLL"
  LangString DESC_SecOpenGL ${LANG_ENGLISH} "Software renderer if your graphics card is missing OpenGL support"

  ;Assign language strings to sections
  !insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} $(DESC_SecCore)
    !insertmacro MUI_DESCRIPTION_TEXT ${SecTech2} $(DESC_SecTech2)
    !insertmacro MUI_DESCRIPTION_TEXT ${SecOpenGL} $(DESC_SecOpenGL)
  !insertmacro MUI_FUNCTION_DESCRIPTION_END
 
;--------------------------------
;Uninstaller Section


!macro DeleteFileOrAskAbort path
  ClearErrors
  Delete "${path}"
  IfErrors 0 +3
    MessageBox MB_ABORTRETRYIGNORE|MB_ICONSTOP 'Unable to delete "${path}"!' IDRETRY -3 IDIGNORE +2
    Abort "Aborted"
!macroend

Section "Uninstall"
  !insertmacro DeleteFileOrAskAbort "$InstDir\LightControl.exe"
  Delete "$InstDir\Uninstall.exe"
  Delete "$InstDir\opengl32.dll"
  Delete "$InstDir\readme.txt"
  RMDir /r "$InstDir\Tech2 SAE J2534 DLL"
  RMDir "$InstDir"
  
  !insertmacro MUI_STARTMENU_GETFOLDER Application $StartMenuFolder
  Delete "$SMPROGRAMS\$StartMenuFolder\LightControl.lnk"  
  Delete "$SMPROGRAMS\$StartMenuFolder\Uninstall.lnk"
  RMDir "$SMPROGRAMS\$StartMenuFolder"
SectionEnd
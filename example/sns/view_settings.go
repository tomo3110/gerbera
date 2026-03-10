package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type SettingsView struct {
	baseView

	settingsEmail       string
	settingsDisplayName string
	settingsBio         string
}

func NewSettingsView(db *sql.DB, hub *Hub) *SettingsView {
	return &SettingsView{baseView: baseView{db: db, hub: hub}}
}

func (v *SettingsView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	v.settingsDisplayName = v.user.DisplayName
	v.settingsEmail = v.user.Email
	v.settingsBio = v.user.Bio
	return nil
}

func (v *SettingsView) Unmount() {
	v.unmountBase()
}

func (v *SettingsView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "settingsDisplayNameInput":
		v.settingsDisplayName = payload["value"]
	case "settingsEmailInput":
		v.settingsEmail = payload["value"]
	case "settingsBioInput":
		v.settingsBio = payload["value"]
	case "saveProfile":
		err := dbUpdateUserProfile(v.db, v.userID, v.settingsDisplayName, v.settingsEmail, v.settingsBio)
		if err != nil {
			v.showToast("Failed to save profile", "danger")
			return nil
		}
		v.user.DisplayName = v.settingsDisplayName
		v.user.Email = v.settingsEmail
		v.user.Bio = v.settingsBio
		v.showToast("Profile saved", "success")
	case "savePassword":
		oldPass := payload["old_password"]
		newPass := payload["new_password"]
		if oldPass == "" || newPass == "" {
			v.showToast("Both fields are required", "warning")
			return nil
		}
		if !verifyPassword(v.user.PasswordHash, oldPass) {
			v.showToast("Current password is incorrect", "danger")
			return nil
		}
		hash := hashPassword(newPass)
		if err := dbUpdateUserPassword(v.db, v.userID, hash); err != nil {
			v.showToast("Failed to update password", "danger")
			return nil
		}
		v.user.PasswordHash = hash
		v.showToast("Password updated", "success")
	case "dismissToast":
		v.toastVisible = false
	case "gerbera:upload_complete":
		// handled in HandleUpload
	}
	return nil
}

func (v *SettingsView) HandleUpload(event string, files []gl.UploadedFile) error {
	if event == "avatarUpload" && len(files) > 0 {
		f := files[0]
		if !strings.HasPrefix(f.MIMEType, "image/") {
			return nil
		}
		ext := ".jpg"
		switch f.MIMEType {
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		}
		filename := fmt.Sprintf("%d%s", v.userID, ext)
		if err := os.WriteFile(filepath.Join("uploads", "avatars", filename), f.Data, 0644); err != nil {
			return err
		}
		avatarPath := "/avatars/" + filename
		if err := dbUpdateUserAvatar(v.db, v.userID, avatarPath); err != nil {
			return err
		}
		v.user.AvatarPath = avatarPath
		v.showToast("Avatar updated", "success")
	}
	return nil
}

func (v *SettingsView) HandleInfo(msg any) error {
	v.handleBaseInfo(msg)
	if _, ok := msg.(NewMessageNotif); ok {
		v.showToast("New message received", "info")
	}
	return nil
}

func (v *SettingsView) Render() g.Components {
	return g.Components{
		gd.Body(
			gd.Div(gp.Attr("style", "padding: var(--g-space-md) 0; font-size: 1.1rem; font-weight: 700"),
				gp.Value("Settings"),
			),
			gu.Card(
				gd.Div(gp.Class("settings-section"),
					gd.H3(gp.Value("Profile")),
					gu.Stack(
						gu.FormGroup(
							gu.FormLabel("Display Name", "display_name"),
							gu.FormInput("display_name",
								gp.ID("display_name"),
								gp.Attr("value", v.settingsDisplayName),
								gl.Input("settingsDisplayNameInput"),
							),
						),
						gu.FormGroup(
							gu.FormLabel("Email", "email"),
							gu.FormInput("email",
								gp.ID("email"),
								gp.Attr("value", v.settingsEmail),
								gp.Attr("type", "email"),
								gl.Input("settingsEmailInput"),
							),
						),
						gu.FormGroup(
							gu.FormLabel("Bio", "bio"),
							gu.FormTextarea("bio",
								gp.Value(v.settingsBio),
								gp.Placeholder("Tell us about yourself"),
								gp.Attr("maxlength", "160"),
								gl.Input("settingsBioInput"),
							),
						),
						gu.Button("Save Profile", gu.ButtonPrimary, gl.Click("saveProfile")),
					),
				),
				gd.Div(gp.Class("settings-section"),
					gd.H3(gp.Value("Change Password")),
					gd.Form(
						gl.Submit("savePassword"),
						gu.Stack(
							gu.FormGroup(
								gu.FormLabel("Current Password", "old_password"),
								gu.FormInput("old_password",
									gp.ID("old_password"),
									gp.Attr("type", "password"),
								),
							),
							gu.FormGroup(
								gu.FormLabel("New Password", "new_password"),
								gu.FormInput("new_password",
									gp.ID("new_password"),
									gp.Attr("type", "password"),
								),
							),
							gu.Button("Update Password", gu.ButtonPrimary),
						),
					),
				),
				gd.Div(gp.Class("settings-section"),
					gd.H3(gp.Value("Avatar")),
					gd.Div(
						gp.Attr("style", "display:flex; align-items:center; gap:var(--g-space-md)"),
						func() g.ComponentFunc {
							if v.user.AvatarPath != "" {
								return gu.ImageAvatar(v.user.AvatarPath, gu.AvatarOpts{Size: "xl"}, gp.Key("avatar-img"))
							}
							return gu.LetterAvatar(v.user.DisplayName, gu.AvatarOpts{Size: "xl"}, gp.Key("avatar-letter"))
						}(),
						gd.Label(
							gp.Attr("style", "cursor:pointer"),
							gd.Input(gp.Type("file"), gl.Upload("avatarUpload"), gl.UploadAccept("image/*"),
								gp.Attr("style", "display:none")),
							gd.Span(gp.Class("g-btn", "g-btn-outline", "g-btn-sm"), gp.Value("Upload Avatar")),
						),
					),
				),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
		),
	}
}

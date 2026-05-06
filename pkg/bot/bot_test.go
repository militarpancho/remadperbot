package bot

import (
	"strings"
	"testing"

	"remadperbot/pkg/models"
	"remadperbot/pkg/scraper"
)

func TestArticleURLUsesCurrentRemadDetailRoute(t *testing.T) {
	got := articleURL("abc123")
	want := "https://remad.madrid.es/REMAD_FTP/#/detalleAntique/abc123"

	if got != want {
		t.Fatalf("articleURL = %q, want %q", got, want)
	}
}

func TestPlanTrackedItemAlertNotifiesStatusChangeWithoutDeletingSubscription(t *testing.T) {
	itemUpdate := models.ItemUpdate{ID: "abc123", Status: "Publicado"}
	articleInfo := &scraper.ArticleInfo{ID: "abc123", Status: "Reservado"}

	plan := planTrackedItemAlert(itemUpdate, articleInfo)

	if !plan.Notify {
		t.Fatalf("Notify = false, want true")
	}
	if plan.Status != "Reservado" {
		t.Fatalf("Status = %q, want Reservado", plan.Status)
	}
	if plan.DeleteSubscription {
		t.Fatalf("DeleteSubscription = true, want false")
	}
}

func TestPlanTrackedItemAlertDeletesSubscriptionWhenProductIsUnavailable(t *testing.T) {
	itemUpdate := models.ItemUpdate{ID: "abc123", Status: "Reservado"}

	plan := planTrackedItemAlert(itemUpdate, nil)

	if !plan.Notify {
		t.Fatalf("Notify = false, want true")
	}
	if plan.Status != "No disponible" {
		t.Fatalf("Status = %q, want No disponible", plan.Status)
	}
	if !plan.DeleteSubscription {
		t.Fatalf("DeleteSubscription = false, want true")
	}
}

func TestNotificationArticleInfoFallsBackToCurrentArticleWhenImageDownloadFails(t *testing.T) {
	currentArticle := &scraper.ArticleInfo{ID: "abc123", Status: "Reservado", Title: "Producto"}

	got := notificationArticleInfo(currentArticle, nil)

	if got != currentArticle {
		t.Fatalf("notificationArticleInfo did not fall back to the current article")
	}
}

func TestTrackedItemNotificationForExistingArticleUsesPhotoAndStatusCaption(t *testing.T) {
	articleInfo := &scraper.ArticleInfo{
		Title:    `<a href="https://remad.example/detalle/abc123">Silla</a>`,
		Metadata: []string{"Categoría: Hogar / Muebles", "Punto limpio: Villaverde", "Descripción: Silla", "Estado: Reservado"},
		Img:      []byte("image bytes"),
	}

	notification := trackedItemNotification(articleInfo)

	if !notification.Photo {
		t.Fatalf("Photo = false, want true")
	}
	if string(notification.PhotoBytes) != "image bytes" {
		t.Fatalf("PhotoBytes = %q, want image bytes", string(notification.PhotoBytes))
	}
	wantCaption := `<a href="https://remad.example/detalle/abc123">Silla</a>` + "\nCambio en el estado del artículo: \nEstado: Reservado"
	if notification.Caption != wantCaption {
		t.Fatalf("Caption = %q, want %q", notification.Caption, wantCaption)
	}
}

func TestTrackedItemNotificationForUnavailableArticleUsesTextOnly(t *testing.T) {
	notification := trackedItemNotification(nil)

	if notification.Photo {
		t.Fatalf("Photo = true, want false")
	}
	if notification.Caption == "" {
		t.Fatalf("Caption is empty")
	}
	if !strings.Contains(notification.Caption, "No disponible") {
		t.Fatalf("Caption = %q, want it to mention No disponible", notification.Caption)
	}
}

func TestTrackedItemNotificationForExistingArticleWithoutImageUsesTextOnly(t *testing.T) {
	articleInfo := &scraper.ArticleInfo{
		Title:    `<a href="https://remad.example/detalle/abc123">Silla</a>`,
		Metadata: []string{"Categoría: Hogar / Muebles", "Punto limpio: Villaverde", "Descripción: Silla", "Estado: Reservado"},
		Status:   "Reservado",
	}

	notification := trackedItemNotification(articleInfo)

	if notification.Photo {
		t.Fatalf("Photo = true, want false")
	}
	if !strings.Contains(notification.Caption, "Estado: Reservado") {
		t.Fatalf("Caption = %q, want it to include current status", notification.Caption)
	}
}

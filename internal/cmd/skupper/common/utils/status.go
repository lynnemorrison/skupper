package utils

import (
	"github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
)

func SiteConfigured(siteList *v1alpha1.SiteList) bool {
	for _, s := range siteList.Items {
		if s.IsConfigured() {
			return true
		}
	}
	return false
}

func SiteReady(siteList *v1alpha1.SiteList) (bool, string) {
	for _, s := range siteList.Items {
		if s.IsReady() {
			return true, s.Name
		}
	}
	return false, ""
}

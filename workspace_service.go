package main

import (
	"changeme/internal/todo"
	"fmt"
	"net/url"
)

func (s *TodoService) Tags() ([]todo.Tag, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().Tags()
}
func (s *TodoService) SaveTags(tags []todo.Tag) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().SaveTags(tags)
}
func (s *TodoService) SetTaskTags(c todo.Command, ids []string) ([]todo.Task, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().SetTaskTags(c, ids)
}
func (s *TodoService) Notes(search string) ([]todo.Note, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().Notes(search)
}
func (s *TodoService) SaveNote(n todo.Note) (todo.Note, error) {
	if err := s.ready(); err != nil {
		return n, err
	}
	return s.core().SaveNote(n)
}
func (s *TodoService) DeleteNote(id string, version int64) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.core().DeleteNote(id, version)
}
func (s *TodoService) TaskReport(q todo.Query) ([]todo.Task, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().TaskReport(q)
}
func (s *TodoService) SetTaskSpan(c todo.Command, start int64, end *int64) ([]todo.Task, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.core().SetTaskSpan(c, start, end)
}

func (s *TodoService) OpenExternal(link string) error {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "mailto" {
		return fmt.Errorf("仅支持 http、https 和邮件链接")
	}
	return s.app.Browser.OpenURL(link)
}

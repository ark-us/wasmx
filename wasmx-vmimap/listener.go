package vmimap

import (
	"context"
	"fmt"
	"log"

	"github.com/emersion/go-imap/v2/imapclient"
)

// StartListener starts an IMAP listener for a session
func (lm *ImapOpenConnection) StartListener(
	goCtx context.Context,
	folder string,
	dataHandler *imapclient.UnilateralDataHandler,
) error {
	// stop previous listener if exists and start a new one
	listener, ok := lm.GetListener(folder)
	if ok && listener != nil {
		lm.StopListener(folder)
		lm.DeleteListener(folder)
	}
	opts := &imapclient.Options{UnilateralDataHandler: dataHandler}
	client, err := lm.GetClient(opts)
	if err != nil {
		return err
	}

	_, err = client.Select(folder, nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select mailbox: %v", err.Error())
	}

	listener = &IMAPListener{Client: client, Folder: folder}
	err = lm.SetListener(folder, listener)
	if err != nil {
		return err
	}

	// Start listening in a goroutine
	go lm.listenForEmails(goCtx, listener)
	return nil
}

func (lm *ImapOpenConnection) StopListener(folder string) {
	listener, ok := lm.GetListener(folder)
	if ok && listener != nil {
		listener.Client.Close()
	}
}

// listenForEmails handles real-time updates using IMAP IDLE
func (lm *ImapOpenConnection) listenForEmails(ctx context.Context, listener *IMAPListener) error {
	var idleCmd *imapclient.IdleCommand
	var err error

	// fmt.Println("* IDLE for incoming emails in: ", listener.Folder)

	errCh := make(chan error, 1)
	go func() {
		idleCmd, err = listener.Client.Idle()
		if err != nil {
			log.Printf("IDLE error: %v \n", err)
			errCh <- err
			return
		}
		err = idleCmd.Wait()
		if err != nil {
			log.Printf("IDLE.Wait error: %v \n", err)
			errCh <- err
			return
		}
		log.Printf("Start IDLE: %s \n", lm.Username)
	}()

	// Wait for the idleDuration or a done signal, whichever comes first.
	select {
	case <-lm.GoContextParent.Done():
		// Received signal to exit idle loop.
		if err := idleCmd.Close(); err != nil {
			log.Printf("Error stopping IDLE: %v \n", err)
		}
		listener.Client.Close()
		return nil
	case <-ctx.Done():
		// Received signal to exit idle loop.
		if err := idleCmd.Close(); err != nil {
			log.Printf("Error stopping IDLE: %v \n", err)
		}
		log.Println("Exiting IMAP idle loop")
		listener.Client.Close()
		return nil
	case err := <-errCh:
		fmt.Println("* IMAP IDLE error", err)
		listener.Client.Close()
		return err
	}
}

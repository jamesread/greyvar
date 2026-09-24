package greyvarEditorTests.gridLoading;

import java.io.File;

import org.junit.rules.ErrorCollector;

import greyvarEditor.EntityDatabase;
import junit.framework.TestCase;

import static org.junit.Assert.*;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.greaterThan;
 
public class TestEntdefs extends TestCase {
	public ErrorCollector collector = new ErrorCollector();
	
	public void testEntdefs() throws NullPointerException {
		EntityDatabase entdb = new EntityDatabase(new File("../../../server/dat/entdefs/"));
			
		entdb.reloadEntdefs(); 
		
		collector.checkThat(entdb.size(), greaterThan(0));
	}
}
 